// File: internal/infrastructure/push/fcm_adapter.go
//
// Firebase Cloud Messaging (HTTP v1) adapteri.
//
// Firebase Admin SDK-ni gətirmek evezine service account JSON-u ile
// ozumuz OAuth2 access token alıriq: RS256 imzali JWT assertion ->
// token endpoint -> access_token (1 saat, kesde saxlanilir).
// Bu, layihede artiq olan golang-jwt-den basqa asililiq teleb etmir.
package push

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/notification"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/golang-jwt/jwt/v5"
)

const (
	fcmScope       = "https://www.googleapis.com/auth/firebase.messaging"
	fcmSendURL     = "https://fcm.googleapis.com/v1/projects/%s/messages:send"
	tokenLifetime  = time.Hour
	tokenSafetyGap = 5 * time.Minute
)

// serviceAccount – Firebase-in verdiyi JSON-un lazim olan hissesi.
type serviceAccount struct {
	Type        string `json:"type"`
	ProjectID   string `json:"project_id"`
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
	TokenURI    string `json:"token_uri"`
}

// FCMAdapter – notification.PushSender realizasiyasi.
type FCMAdapter struct {
	account *serviceAccount
	client  *http.Client
	log     logger.Logger

	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

// NewFCMAdapter – konfiqurasiya yoxdursa (nil olmayan) sondurulmus
// adapter qaytarir: Enabled() false olur, Send() sessizce kecir.
//
// credentialsPath – service account JSON faylinin yolu
// credentialsJSON – eyni JSON-un birbasa deyeri (env-de saxlamaq ucun)
func NewFCMAdapter(credentialsPath, credentialsJSON string, log logger.Logger) *FCMAdapter {
	adapter := &FCMAdapter{
		client: &http.Client{Timeout: 15 * time.Second},
		log:    log,
	}

	raw := strings.TrimSpace(credentialsJSON)
	if raw == "" && strings.TrimSpace(credentialsPath) != "" {
		content, err := os.ReadFile(credentialsPath)
		if err != nil {
			log.Warn("FCM credentials faylı oxunmadi, push sondurulub",
				logger.Field{Key: "path", Value: credentialsPath},
				logger.Field{Key: "error", Value: err.Error()},
			)
			return adapter
		}
		raw = string(content)
	}

	if raw == "" {
		log.Info("FCM konfiqurasiyasi yoxdur, push sondurulub")
		return adapter
	}

	var account serviceAccount
	if err := json.Unmarshal([]byte(raw), &account); err != nil {
		log.Warn("FCM credentials JSON parse olunmadi, push sondurulub",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return adapter
	}

	if account.ProjectID == "" || account.ClientEmail == "" || account.PrivateKey == "" {
		log.Warn("FCM credentials natamamdir, push sondurulub")
		return adapter
	}

	if account.TokenURI == "" {
		account.TokenURI = "https://oauth2.googleapis.com/token"
	}

	adapter.account = &account
	log.Info("FCM push aktivdir",
		logger.Field{Key: "project_id", Value: account.ProjectID},
	)
	return adapter
}

// Enabled – konfiqurasiya varmi?
func (a *FCMAdapter) Enabled() bool {
	return a.account != nil
}

// Send – bir bildirisi butun token-lere gonderir.
//
// FCM v1 batch endpoint-i deprecate olunub, ona gore token-ler bir-bir
// gonderilir. Bir token-in xetasi (mes. artiq etibarsizdir) qalanlarini
// dayandirmir; hamisi ugursuz olarsa error qaytarilir ki, outbox
// yeniden cehd etsin.
func (a *FCMAdapter) Send(ctx context.Context, tokens []string, envelope *notification.Envelope) error {
	if !a.Enabled() || len(tokens) == 0 {
		return nil
	}

	accessToken, err := a.ensureAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("fcm access token alinmadi: %w", err)
	}

	endpoint := fmt.Sprintf(fcmSendURL, a.account.ProjectID)
	data := buildDataPayload(envelope)

	var lastErr error
	delivered := 0

	for _, token := range tokens {
		message := map[string]interface{}{
			"message": map[string]interface{}{
				"token": token,
				"notification": map[string]interface{}{
					"title": envelope.Title,
					"body":  envelope.Body,
				},
				"data": data,
				"android": map[string]interface{}{
					"priority": "high",
				},
				"apns": map[string]interface{}{
					"headers": map[string]string{"apns-priority": "10"},
				},
			},
		}

		if err := a.postMessage(ctx, endpoint, accessToken, message); err != nil {
			lastErr = err
			a.log.Warn("FCM gonderisi ugursuz",
				logger.Field{Key: "error", Value: err.Error()},
			)
			continue
		}
		delivered++
	}

	if delivered == 0 && lastErr != nil {
		return lastErr
	}
	return nil
}

func (a *FCMAdapter) postMessage(
	ctx context.Context,
	endpoint, accessToken string,
	message map[string]interface{},
) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("fcm mesaji serializasiya olunmadi: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := a.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}

	detail, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
	return fmt.Errorf("fcm status %d: %s", response.StatusCode, strings.TrimSpace(string(detail)))
}

// ensureAccessToken – kesde etibarli token varsa onu qaytarir,
// yoxsa yenisini alir.
func (a *FCMAdapter) ensureAccessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.accessToken != "" && time.Now().Before(a.expiresAt.Add(-tokenSafetyGap)) {
		return a.accessToken, nil
	}

	assertion, err := a.buildAssertion()
	if err != nil {
		return "", err
	}

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", assertion)

	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, a.account.TokenURI, strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := a.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return "", fmt.Errorf("token endpoint status %d: %s", response.StatusCode, strings.TrimSpace(string(detail)))
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", fmt.Errorf("token endpoint bos access_token qaytardi")
	}

	a.accessToken = payload.AccessToken
	a.expiresAt = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	if payload.ExpiresIn == 0 {
		a.expiresAt = time.Now().Add(tokenLifetime)
	}

	return a.accessToken, nil
}

// buildAssertion – service account ile imzalanmis JWT.
func (a *FCMAdapter) buildAssertion() (string, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(a.account.PrivateKey))
	if err != nil {
		return "", fmt.Errorf("fcm private key parse olunmadi: %w", err)
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   a.account.ClientEmail,
		"scope": fcmScope,
		"aud":   a.account.TokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(tokenLifetime).Unix(),
	})

	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("fcm assertion imzalanmadi: %w", err)
	}
	return signed, nil
}

// buildDataPayload – FCM data saheleri yalniz string qebul edir,
// ona gore ic-ice obyektler JSON string kimi gonderilir.
func buildDataPayload(envelope *notification.Envelope) map[string]string {
	data := map[string]string{
		"type": string(envelope.Type),
	}

	if envelope.BookingID != nil {
		data["booking_id"] = envelope.BookingID.String()
	}
	if envelope.BusinessID != nil {
		data["business_id"] = envelope.BusinessID.String()
	}

	for key, value := range envelope.Payload {
		switch typed := value.(type) {
		case string:
			data[key] = typed
		case nil:
			continue
		default:
			if encoded, err := json.Marshal(typed); err == nil {
				data[key] = string(encoded)
			}
		}
	}

	return data
}
