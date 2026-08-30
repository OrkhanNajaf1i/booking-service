// File: internal/domain/availability/service_test.go
package availability

import (
	"context"
	"testing"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

// ============================================================
// TEST KOMEKCILERI
// ============================================================

// fakeRepo – yaddasda isleyen sade repository.
type fakeRepo struct {
	workingHours []*WorkingHours
	settings     *ScheduleSettings
	booked       []BusyInterval
	timeOff      []*TimeOff
}

func (f *fakeRepo) UpsertWorkingHours(context.Context, *WorkingHours) error { return nil }

func (f *fakeRepo) ListWorkingHours(context.Context, uuid.UUID, uuid.UUID) ([]*WorkingHours, error) {
	return f.workingHours, nil
}

func (f *fakeRepo) GetWorkingHoursForDay(_ context.Context, _, _ uuid.UUID, day int) (*WorkingHours, error) {
	for _, wh := range f.workingHours {
		if wh.DayOfWeek == day {
			return wh, nil
		}
	}
	return nil, nil
}

func (f *fakeRepo) DeleteWorkingHours(context.Context, uuid.UUID, uuid.UUID, int) error { return nil }

func (f *fakeRepo) GetSettings(context.Context, uuid.UUID, *uuid.UUID) (*ScheduleSettings, error) {
	return f.settings, nil
}

func (f *fakeRepo) UpsertSettings(context.Context, *ScheduleSettings) error { return nil }

func (f *fakeRepo) CreateTimeOff(context.Context, *TimeOff) error { return nil }

func (f *fakeRepo) ListTimeOff(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time) ([]*TimeOff, error) {
	return f.timeOff, nil
}

func (f *fakeRepo) DeleteTimeOff(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func (f *fakeRepo) ListBookedIntervals(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time) ([]BusyInterval, error) {
	return f.booked, nil
}

// nopLogger – testde log cixisini susdurur.
type nopLogger struct{}

func (nopLogger) Info(string, ...logger.Field)  {}
func (nopLogger) Debug(string, ...logger.Field) {}
func (nopLogger) Error(string, ...logger.Field) {}
func (nopLogger) Warn(string, ...logger.Field)  {}

// testDay – Cume axsami, 2026-09-03. Sabit tarix secilib ki,
// heftenin gunu deyismesin.
var testDay = time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)

func newTestService(repo *fakeRepo) *availabilityService {
	service := NewService(repo, nil, nopLogger{}).(*availabilityService)
	// Vaxti sabitleyirik: test gunundan bir gun evvel, gece yarisi.
	service.now = func() time.Time {
		return time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	}
	return service
}

func baseSettings(stepMins, durationMins int) *ScheduleSettings {
	settings := DefaultSettings(uuid.New())
	settings.Timezone = "UTC"
	settings.SlotStepMins = stepMins
	settings.DefaultDurationMins = durationMins
	settings.MinNoticeMins = 0
	settings.MaxAdvanceDays = 30
	return settings
}

func nineToSix(breakEnabled bool) *WorkingHours {
	breakStart := "13:00"
	breakEnd := "14:00"
	return &WorkingHours{
		ID:           uuid.New(),
		DayOfWeek:    int(testDay.Weekday()),
		StartTime:    "09:00",
		EndTime:      "18:00",
		BreakEnabled: breakEnabled,
		BreakStart:   &breakStart,
		BreakEnd:     &breakEnd,
		IsActive:     true,
	}
}

func availableStarts(day DayAvailability) []string {
	starts := make([]string, 0, len(day.Slots))
	for _, slot := range day.Slots {
		if slot.Available {
			starts = append(starts, slot.Start.Format("15:04"))
		}
	}
	return starts
}

func computeDay(t *testing.T, repo *fakeRepo) DayAvailability {
	t.Helper()

	result, err := newTestService(repo).GetAvailability(context.Background(), uuid.New(), &AvailabilityQuery{
		StaffID:  uuid.New(),
		FromDate: testDay,
		ToDate:   testDay,
	})
	if err != nil {
		t.Fatalf("GetAvailability xetasi: %v", err)
	}
	if len(result.Days) != 1 {
		t.Fatalf("1 gun gozlenilirdi, %d geldi", len(result.Days))
	}
	return result.Days[0]
}

// ============================================================
// TESTLER
// ============================================================

// 09:00–18:00, 60 deqiqelik addim, fasile yoxdur -> 9 slot.
func TestHourlyStepWithoutBreak(t *testing.T) {
	day := computeDay(t, &fakeRepo{
		workingHours: []*WorkingHours{nineToSix(false)},
		settings:     baseSettings(60, 60),
	})

	got := availableStarts(day)
	want := []string{"09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00"}

	if len(got) != len(want) {
		t.Fatalf("slot sayi: gozlenilen %d, gelen %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("slot %d: gozlenilen %s, gelen %s", i, want[i], got[i])
		}
	}
}

// Nahar fasilesi aktiv olanda 13:00 slotu itmelidir.
func TestLunchBreakRemovesSlots(t *testing.T) {
	day := computeDay(t, &fakeRepo{
		workingHours: []*WorkingHours{nineToSix(true)},
		settings:     baseSettings(60, 60),
	})

	if day.Break == nil {
		t.Fatal("gunun fasile melumati gozlenilirdi")
	}
	if day.Break.Start != "13:00" || day.Break.End != "14:00" {
		t.Errorf("fasile 13:00-14:00 gozlenilirdi, %s-%s geldi", day.Break.Start, day.Break.End)
	}

	for _, start := range availableStarts(day) {
		if start == "13:00" {
			t.Error("nahar fasilesinde slot verilmemelidir")
		}
	}

	// 09:00-13:00 -> 4 slot, 14:00-18:00 -> 4 slot
	if got := len(availableStarts(day)); got != 8 {
		t.Errorf("8 slot gozlenilirdi, %d geldi: %v", got, availableStarts(day))
	}
}

// Admin addimi 16 deqiqe qoyanda vaxtlar 16 deqiqe araliqla gelmelidir.
func TestSixteenMinuteStep(t *testing.T) {
	day := computeDay(t, &fakeRepo{
		workingHours: []*WorkingHours{nineToSix(false)},
		settings:     baseSettings(16, 16),
	})

	starts := availableStarts(day)
	want := []string{"09:00", "09:16", "09:32", "09:48", "10:04"}

	for i, expected := range want {
		if starts[i] != expected {
			t.Errorf("slot %d: gozlenilen %s, gelen %s", i, expected, starts[i])
		}
	}

	// 09:00–18:00 = 540 deqiqe; 16 deqiqelik randevu 33 defe yerlesir
	// (33*16 = 528 <= 540, 34*16 = 544 > 540).
	if len(starts) != 33 {
		t.Errorf("33 slot gozlenilirdi, %d geldi", len(starts))
	}
}

// Movcud bron oz vaxtini tutmalidir, qonsu slotlar acıq qalmalidir.
func TestExistingBookingBlocksItsSlot(t *testing.T) {
	booked := BusyInterval{
		Start: time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 9, 3, 11, 0, 0, 0, time.UTC),
	}

	day := computeDay(t, &fakeRepo{
		workingHours: []*WorkingHours{nineToSix(false)},
		settings:     baseSettings(60, 60),
		booked:       []BusyInterval{booked},
	})

	for _, start := range availableStarts(day) {
		if start == "10:00" {
			t.Error("bron edilmis vaxt acıq gorunmemelidir")
		}
	}

	// Serhedde toxunan slotlar bloklanmamalidir.
	for _, expected := range []string{"09:00", "11:00"} {
		found := false
		for _, start := range availableStarts(day) {
			if start == expected {
				found = true
			}
		}
		if !found {
			t.Errorf("%s slotu acıq qalmalidir", expected)
		}
	}
}

// Bufer movcud randevunun her iki terefinde bosluq saxlamalidir.
func TestBufferBlocksNeighbouringSlots(t *testing.T) {
	settings := baseSettings(60, 60)
	settings.BufferAfterMins = 15

	booked := BusyInterval{
		Start: time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 9, 3, 11, 0, 0, 0, time.UTC),
	}

	day := computeDay(t, &fakeRepo{
		workingHours: []*WorkingHours{nineToSix(false)},
		settings:     settings,
		booked:       []BusyInterval{booked},
	})

	// 11:00 artiq 15 deqiqelik buferə dusur.
	for _, start := range availableStarts(day) {
		if start == "11:00" {
			t.Error("bufer sebebinden 11:00 acıq olmamalidir")
		}
	}
	// 12:00 bufer sahesinden kenardadir.
	found := false
	for _, start := range availableStarts(day) {
		if start == "12:00" {
			found = true
		}
	}
	if !found {
		t.Error("12:00 slotu acıq qalmalidir")
	}
}

// Is gunu olmayan gun ucun slot verilmemelidir.
func TestNonWorkingDayHasNoSlots(t *testing.T) {
	otherDay := nineToSix(false)
	otherDay.DayOfWeek = (int(testDay.Weekday()) + 1) % 7

	day := computeDay(t, &fakeRepo{
		workingHours: []*WorkingHours{otherDay},
		settings:     baseSettings(60, 60),
	})

	if day.IsWorkday {
		t.Error("bu gun is gunu olmamalidir")
	}
	if len(day.Slots) != 0 {
		t.Errorf("slot gozlenilmirdi, %d geldi", len(day.Slots))
	}
}

// min_notice penceresine dusen slotlar secile bilmez.
func TestMinNoticeMarksSlotsTooSoon(t *testing.T) {
	settings := baseSettings(60, 60)
	settings.MinNoticeMins = 120

	repo := &fakeRepo{
		workingHours: []*WorkingHours{nineToSix(false)},
		settings:     settings,
	}

	service := NewService(repo, nil, nopLogger{}).(*availabilityService)
	// Indi test gunu saat 09:30-dur.
	service.now = func() time.Time {
		return time.Date(2026, 9, 3, 9, 30, 0, 0, time.UTC)
	}

	result, err := service.GetAvailability(context.Background(), uuid.New(), &AvailabilityQuery{
		StaffID:  uuid.New(),
		FromDate: testDay,
		ToDate:   testDay,
	})
	if err != nil {
		t.Fatalf("GetAvailability xetasi: %v", err)
	}

	states := map[string]SlotState{}
	for _, slot := range result.Days[0].Slots {
		states[slot.Start.Format("15:04")] = slot.State
	}

	if states["09:00"] != SlotStatePast {
		t.Errorf("09:00 past olmalidir, %s geldi", states["09:00"])
	}
	if states["10:00"] != SlotStateTooSoon {
		t.Errorf("10:00 too_soon olmalidir, %s geldi", states["10:00"])
	}
	if states["12:00"] != SlotStateAvailable {
		t.Errorf("12:00 available olmalidir, %s geldi", states["12:00"])
	}
}

// CheckSlot yalniz grid uzerindeki bos vaxti qebul etmelidir.
func TestCheckSlotRejectsOffGridTime(t *testing.T) {
	repo := &fakeRepo{
		workingHours: []*WorkingHours{nineToSix(false)},
		settings:     baseSettings(60, 60),
	}
	service := newTestService(repo)

	onGrid := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	if _, err := service.CheckSlot(context.Background(), uuid.New(), uuid.New(), nil, onGrid); err != nil {
		t.Errorf("grid uzerindeki vaxt qebul edilmeliydi: %v", err)
	}

	offGrid := time.Date(2026, 9, 3, 10, 17, 0, 0, time.UTC)
	if _, err := service.CheckSlot(context.Background(), uuid.New(), uuid.New(), nil, offGrid); err == nil {
		t.Error("grid xaricindeki vaxt redd edilmeliydi")
	}
}
