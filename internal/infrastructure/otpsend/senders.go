// File: internal/infrastructure/otpsend/senders.go
//
// Kod catdiran kanallar.
//
// Domen yalniz "kod gonderilsin" deyir; hansi xidmetin islendiyi
// buradadir. Provayder deyisende yalniz bu fayl toxunulur.
//
// PULSUZ variant yoxdur — real SMS her halda pulludur. Ona gore:
//
//   - `log`      inkisaf ve test: kod jurnala yazilir, xarici xidmet
//     yoxdur, tamamile pulsuzdur
//   - `whatsapp` Meta Cloud API: SMS-den ucuzdur, amma "authentication"
//     sablonu yene odenislidir. Test nomresi ile 5 tesdiq
//     edilmis alicIya pulsuz gondermek olur
//   - `sms`      istenilen HTTP API-li operator; her mesaj odenislidir
package otpsend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/otp"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
)

// ── Jurnal kanali ────────────────────────────────────────────

// LogSender – kodu yalniz jurnala yazir.
//
// Inkisaf ucundur: SMS provayderi qurulmamis butun axini yoxlamaq
// olur. Xidmet `sms`/`whatsapp` acarlari olmadan qalxanda avtomatik
// bu secilir.
type LogSender struct {
	log logger.Logger
}

func NewLogSender(log logger.Logger) *LogSender {
	return &LogSender{log: log}
}

func (s *LogSender) Send(_ context.Context, phoneE164, code string) error {
	// Nomre maskalanir, kod isə acıq yazilir — inkisaf rejimi
	// oldugu ucun bu qesdendir.
	s.log.Info("TESDIQ KODU (yalniz inkisaf rejimi)",
		logger.Field{Key: "phone", Value: otp.MaskPhone(phoneE164)},
		logger.Field{Key: "code", Value: code},
	)
	return nil
}

func (s *LogSender) Channel() otp.Channel { return otp.ChannelLog }

// ── WhatsApp (Meta Cloud API) ────────────────────────────────

// WhatsAppConfig – Meta Cloud API acarlari.
type WhatsAppConfig struct {
	// PhoneNumberID – Meta paneldeki gonderen nomrenin ID-si.
	PhoneNumberID string
	// AccessToken – daimi (system user) token.
	AccessToken string
	// TemplateName – "authentication" tipli sablonun adi.
	TemplateName string
	// LanguageCode – sablonun dili, mes. "az" ve ya "en".
	LanguageCode string
}

func (c WhatsAppConfig) IsComplete() bool {
	return c.PhoneNumberID != "" && c.AccessToken != "" && c.TemplateName != ""
}

// WhatsAppSender – Meta Cloud API uzerinden tesdiq mesaji.
//
// Meta sərbəst mətnli mesaja icaze vermir: tesdiq kodu yalniz
// "authentication" tipli TESDIQLENMIS sablonla gonderile bilir.
// Kod sablonun deyisen hissesine qoyulur.
type WhatsAppSender struct {
	config WhatsAppConfig
	client *http.Client
	log    logger.Logger
}

func NewWhatsAppSender(config WhatsAppConfig, log logger.Logger) *WhatsAppSender {
	if config.LanguageCode == "" {
		config.LanguageCode = "az"
	}

	return &WhatsAppSender{
		config: config,
		client: &http.Client{Timeout: 15 * time.Second},
		log:    log,
	}
}

func (s *WhatsAppSender) Channel() otp.Channel { return otp.ChannelWhatsApp }

func (s *WhatsAppSender) Send(ctx context.Context, phoneE164, code string) error {
	endpoint := fmt.Sprintf(
		"https://graph.facebook.com/v21.0/%s/messages",
		s.config.PhoneNumberID,
	)

	// WhatsApp nomreni "+" olmadan isteyir.
	recipient := strings.TrimPrefix(phoneE164, "+")

	payload := map[string]any{
		"messaging_product": "whatsapp",
		"to":                recipient,
		"type":              "template",
		"template": map[string]any{
			"name":     s.config.TemplateName,
			"language": map[string]string{"code": s.config.LanguageCode},
			"components": []any{
				// Sablonun govdesindeki {{1}} — kodun ozu.
				map[string]any{
					"type": "body",
					"parameters": []any{
						map[string]string{"type": "text", "text": code},
					},
				},
				// Authentication sablonunda "kopyala" duymesi olur;
				// o da eyni kodu alir.
				map[string]any{
					"type":     "button",
					"sub_type": "url",
					"index":    "0",
					"parameters": []any{
						map[string]string{"type": "text", "text": code},
					},
				},
			},
		},
	}

	return s.post(ctx, endpoint, payload, phoneE164)
}

func (s *WhatsAppSender) post(
	ctx context.Context,
	endpoint string,
	payload map[string]any,
	phoneE164 string,
) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("whatsapp: failed to encode payload: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, endpoint, bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("whatsapp: failed to build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+s.config.AccessToken)

	response, err := s.client.Do(request)
	if err != nil {
		return fmt.Errorf("whatsapp: request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		// Cavab govdesi jurnala dusur, amma nomre maskalanir.
		detail, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		s.log.Error("WhatsApp kod gonderilmedi",
			logger.Field{Key: "phone", Value: otp.MaskPhone(phoneE164)},
			logger.Field{Key: "status", Value: response.StatusCode},
			logger.Field{Key: "detail", Value: string(detail)},
		)
		return fmt.Errorf("whatsapp: unexpected status %d", response.StatusCode)
	}

	return nil
}

// ── SMS (umumi HTTP provayder) ───────────────────────────────

// SMSConfig – istenilen HTTP API-li operator ucun.
//
// Sablonlar sadedir: `{phone}` ve `{message}` evez olunur. Belelikle
// yerli operatorlarin cox hissesi ucun kod yazmaga ehtiyac qalmir.
type SMSConfig struct {
	// Endpoint – POST unvani. Numune:
	//   https://api.operator.az/v1/send
	Endpoint string
	// AuthHeader – "Authorization: Bearer xxx" kimi tam basliq.
	AuthHeader string
	// BodyTemplate – JSON govdesi. Numune:
	//   {"to":"{phone}","text":"{message}","sender":"BOOKIFY"}
	BodyTemplate string
	// MessageTemplate – mesajin metni; `{code}` evez olunur.
	MessageTemplate string
}

func (c SMSConfig) IsComplete() bool {
	return c.Endpoint != "" && c.BodyTemplate != ""
}

type SMSSender struct {
	config SMSConfig
	client *http.Client
	log    logger.Logger
}

func NewSMSSender(config SMSConfig, log logger.Logger) *SMSSender {
	if config.MessageTemplate == "" {
		config.MessageTemplate = "Bookify təsdiq kodunuz: {code}. Kodu heç kimlə paylaşmayın."
	}

	return &SMSSender{
		config: config,
		client: &http.Client{Timeout: 15 * time.Second},
		log:    log,
	}
}

func (s *SMSSender) Channel() otp.Channel { return otp.ChannelSMS }

func (s *SMSSender) Send(ctx context.Context, phoneE164, code string) error {
	message := strings.ReplaceAll(s.config.MessageTemplate, "{code}", code)

	body := s.config.BodyTemplate
	body = strings.ReplaceAll(body, "{phone}", phoneE164)
	body = strings.ReplaceAll(body, "{message}", jsonEscape(message))

	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, s.config.Endpoint, strings.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("sms: failed to build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	if s.config.AuthHeader != "" {
		name, value, found := strings.Cut(s.config.AuthHeader, ":")
		if found {
			request.Header.Set(strings.TrimSpace(name), strings.TrimSpace(value))
		}
	}

	response, err := s.client.Do(request)
	if err != nil {
		return fmt.Errorf("sms: request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		s.log.Error("SMS kod gonderilmedi",
			logger.Field{Key: "phone", Value: otp.MaskPhone(phoneE164)},
			logger.Field{Key: "status", Value: response.StatusCode},
			logger.Field{Key: "detail", Value: string(detail)},
		)
		return fmt.Errorf("sms: unexpected status %d", response.StatusCode)
	}

	return nil
}

// jsonEscape – mesaj JSON sablonunun icine qoyulur, ona gore
// dırnaq və xüsusi simvollar qacirilmalidir.
func jsonEscape(text string) string {
	encoded, err := json.Marshal(text)
	if err != nil {
		return text
	}
	// Marshal dırnaqlarla qaytarir; sablonda dırnaq onsuz da var.
	return string(encoded[1 : len(encoded)-1])
}
