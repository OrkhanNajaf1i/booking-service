package otp

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ── Komekci saxta tetbiqler ──────────────────────────────────

type fakeRepo struct {
	items []*Verification
}

func (r *fakeRepo) Create(_ context.Context, v *Verification) error {
	r.items = append(r.items, v)
	return nil
}

func (r *fakeRepo) LatestActive(_ context.Context, phone string) (*Verification, error) {
	for i := len(r.items) - 1; i >= 0; i-- {
		if r.items[i].PhoneE164 == phone && r.items[i].ConsumedAt == nil {
			return r.items[i], nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) CountSince(_ context.Context, phone string, since time.Time) (int, error) {
	count := 0
	for _, item := range r.items {
		if item.PhoneE164 == phone && item.CreatedAt.After(since) {
			count++
		}
	}
	return count, nil
}

func (r *fakeRepo) IncrementAttempts(_ context.Context, id uuid.UUID) error {
	for _, item := range r.items {
		if item.ID == id {
			item.Attempts++
		}
	}
	return nil
}

func (r *fakeRepo) MarkConsumed(_ context.Context, id uuid.UUID) error {
	for _, item := range r.items {
		if item.ID == id {
			now := time.Now()
			item.ConsumedAt = &now
		}
	}
	return nil
}

type fakeSender struct {
	lastCode  string
	lastPhone string
	sent      int
	failWith  error
}

func (s *fakeSender) Send(_ context.Context, phone, code string) error {
	if s.failWith != nil {
		return s.failWith
	}
	s.lastPhone = phone
	s.lastCode = code
	s.sent++
	return nil
}

func (s *fakeSender) Channel() Channel { return ChannelSMS }

type fixedClock struct{ now time.Time }

func (c *fixedClock) Now() time.Time { return c.now }

func newService(policy Policy) (*Service, *fakeRepo, *fakeSender, *fixedClock) {
	repo := &fakeRepo{}
	sender := &fakeSender{}
	clock := &fixedClock{now: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)}
	return NewService(repo, sender, policy).WithClock(clock), repo, sender, clock
}

// ── Nomre normallasdirmasi ───────────────────────────────────

func TestNormalizePhone(t *testing.T) {
	// Eyni nomrenin butun yazilislari bir setre dusmelidir — eks
	// halda hemin adam her defe yeni hesab yaradar.
	same := []string{
		"0501112233",
		"050 111 22 33",
		"+994 50 111 22 33",
		"994501112233",
		"(050) 111-22-33",
		"501112233",
	}

	for _, raw := range same {
		got, err := NormalizePhone(raw)
		if err != nil {
			t.Fatalf("NormalizePhone(%q) xeta verdi: %v", raw, err)
		}
		if got != "+994501112233" {
			t.Fatalf("NormalizePhone(%q) = %q, gozlenilen +994501112233", raw, got)
		}
	}

	invalid := []string{
		"",
		"123",
		"05011122",         // qisa
		"05011122334",      // uzun
		"+7 900 111 22 33", // basqa olke
		"0301112233",       // taninmayan operator
		"salam",
	}

	for _, raw := range invalid {
		if _, err := NormalizePhone(raw); err == nil {
			t.Fatalf("NormalizePhone(%q) qebul etdi, redd etmeli idi", raw)
		}
	}
}

func TestMaskPhoneHidesMiddle(t *testing.T) {
	masked := MaskPhone("+994501112233")
	if masked == "+994501112233" {
		t.Fatal("nomre maskalanmayib")
	}
	// Son iki reqem gorunmelidir ki, adam oz nomresini taniyasin.
	if masked[len(masked)-2:] != "33" {
		t.Fatalf("son reqemler gorunmur: %q", masked)
	}
	// Operator kodu da gorunur.
	if !strings.Contains(masked, "50") {
		t.Fatalf("operator kodu gorunmur: %q", masked)
	}
	// Ortadaki reqemler gizlenmelidir.
	if strings.Contains(masked, "1112") {
		t.Fatalf("nomrenin ortasi sizib: %q", masked)
	}
}

// ── Kod axini ────────────────────────────────────────────────

func TestRequestAndVerify(t *testing.T) {
	service, _, sender, _ := newService(DefaultPolicy())
	ctx := context.Background()

	result, err := service.RequestCode(ctx, "050 111 22 33")
	if err != nil {
		t.Fatalf("RequestCode: %v", err)
	}
	if result.PhoneE164 != "+994501112233" {
		t.Fatalf("normallasdirma alinmadi: %q", result.PhoneE164)
	}
	if sender.sent != 1 {
		t.Fatalf("kod gonderilmedi, sent=%d", sender.sent)
	}
	if len(sender.lastCode) != 6 {
		t.Fatalf("kod 6 reqemli olmalidir: %q", sender.lastCode)
	}
	// SMS kanalinda kod cavabda GORUNMEMELIDIR.
	if result.DebugCode != "" {
		t.Fatal("real kanalda kod cavaba dusub")
	}

	phone, err := service.VerifyCode(ctx, "0501112233", sender.lastCode)
	if err != nil {
		t.Fatalf("VerifyCode: %v", err)
	}
	if phone != "+994501112233" {
		t.Fatalf("VerifyCode = %q", phone)
	}
}

func TestCodeIsSingleUse(t *testing.T) {
	service, _, sender, _ := newService(DefaultPolicy())
	ctx := context.Background()

	_, _ = service.RequestCode(ctx, "0501112233")
	code := sender.lastCode

	if _, err := service.VerifyCode(ctx, "0501112233", code); err != nil {
		t.Fatalf("birinci yoxlama: %v", err)
	}

	// Ikinci defe islemamelidir: tutulmus kod tekrar oynadila bilmez.
	_, err := service.VerifyCode(ctx, "0501112233", code)
	if err == nil {
		t.Fatal("istifade edilmis kod ikinci defe qebul olundu")
	}
	if otpErr, ok := AsError(err); !ok || otpErr.Code != "CODE_NOT_FOUND" {
		t.Fatalf("gozlenilen CODE_NOT_FOUND, alinan: %v", err)
	}
}

func TestWrongCodeCountsAttempts(t *testing.T) {
	policy := DefaultPolicy()
	policy.MaxAttempts = 3
	service, repo, sender, _ := newService(policy)
	ctx := context.Background()

	_, _ = service.RequestCode(ctx, "0501112233")

	for i := 0; i < 3; i++ {
		if _, err := service.VerifyCode(ctx, "0501112233", "000000"); err == nil {
			t.Fatal("yanlis kod qebul olundu")
		}
	}

	// Limit dolandan sonra DOGRU kod da qebul olunmamalidir.
	_, err := service.VerifyCode(ctx, "0501112233", sender.lastCode)
	if otpErr, ok := AsError(err); !ok || otpErr.Code != "TOO_MANY_ATTEMPTS" {
		t.Fatalf("gozlenilen TOO_MANY_ATTEMPTS, alinan: %v", err)
	}
	if repo.items[0].Attempts != 3 {
		t.Fatalf("cehd sayi yazilmayib: %d", repo.items[0].Attempts)
	}
}

func TestExpiredCodeRejected(t *testing.T) {
	service, _, sender, clock := newService(DefaultPolicy())
	ctx := context.Background()

	_, _ = service.RequestCode(ctx, "0501112233")

	clock.now = clock.now.Add(6 * time.Minute)

	_, err := service.VerifyCode(ctx, "0501112233", sender.lastCode)
	if otpErr, ok := AsError(err); !ok || otpErr.Code != "CODE_EXPIRED" {
		t.Fatalf("gozlenilen CODE_EXPIRED, alinan: %v", err)
	}
}

func TestResendCooldown(t *testing.T) {
	service, _, _, clock := newService(DefaultPolicy())
	ctx := context.Background()

	if _, err := service.RequestCode(ctx, "0501112233"); err != nil {
		t.Fatalf("birinci sorgu: %v", err)
	}

	// Derhal ikinci sorgu bloklanmalidir.
	_, err := service.RequestCode(ctx, "0501112233")
	if otpErr, ok := AsError(err); !ok || otpErr.Code != "RESEND_TOO_SOON" {
		t.Fatalf("gozlenilen RESEND_TOO_SOON, alinan: %v", err)
	}

	clock.now = clock.now.Add(61 * time.Second)
	if _, err := service.RequestCode(ctx, "0501112233"); err != nil {
		t.Fatalf("gozlemeden sonra: %v", err)
	}
}

func TestHourlyLimit(t *testing.T) {
	policy := DefaultPolicy()
	policy.ResendAfter = 0 // burada yalniz saatliq limit yoxlanilir
	service, _, _, clock := newService(policy)
	ctx := context.Background()

	for i := 0; i < policy.MaxPerHour; i++ {
		if _, err := service.RequestCode(ctx, "0501112233"); err != nil {
			t.Fatalf("%d-ci sorgu: %v", i+1, err)
		}
		clock.now = clock.now.Add(time.Second)
	}

	_, err := service.RequestCode(ctx, "0501112233")
	if otpErr, ok := AsError(err); !ok || otpErr.Code != "TOO_MANY_REQUESTS" {
		t.Fatalf("gozlenilen TOO_MANY_REQUESTS, alinan: %v", err)
	}

	// Bir saat sonra yeniden mumkun olmalidir.
	clock.now = clock.now.Add(time.Hour + time.Second)
	if _, err := service.RequestCode(ctx, "0501112233"); err != nil {
		t.Fatalf("bir saat sonra: %v", err)
	}
}

func TestSendFailureDoesNotConsumeLimit(t *testing.T) {
	repo := &fakeRepo{}
	sender := &fakeSender{failWith: context.DeadlineExceeded}
	clock := &fixedClock{now: time.Now()}
	service := NewService(repo, sender, DefaultPolicy()).WithClock(clock)

	if _, err := service.RequestCode(context.Background(), "0501112233"); err == nil {
		t.Fatal("gonderme alinmadi, amma xeta qayitmadi")
	}

	// Gonderilmeyen kod yazilmamalidir — eks halda istifadeci
	// heç bir kod almadan saatliq limitini itirerdi.
	if len(repo.items) != 0 {
		t.Fatalf("gonderilmeyen kod yazilib: %d", len(repo.items))
	}
}

func TestLogChannelReturnsCodeForDevelopment(t *testing.T) {
	repo := &fakeRepo{}
	sender := &logChannelSender{}
	service := NewService(repo, sender, DefaultPolicy())

	result, err := service.RequestCode(context.Background(), "0501112233")
	if err != nil {
		t.Fatalf("RequestCode: %v", err)
	}
	if result.DebugCode == "" {
		t.Fatal("jurnal kanalinda kod cavabda olmalidir")
	}
}

type logChannelSender struct{}

func (logChannelSender) Send(context.Context, string, string) error { return nil }
func (logChannelSender) Channel() Channel                           { return ChannelLog }
