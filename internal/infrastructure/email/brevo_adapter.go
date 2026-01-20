// File: internal/infrastructure/email/brevo_adapter.go
package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// BrevoAdapter - Brevo HTTP API üzərindən mail göndərən adapter.
// Bu struct sadəcə auth.EmailService interfeysindəki metodu implement edir.
type BrevoAdapter struct {
	apiKey      string
	senderEmail string
	httpClient  *http.Client
}

// NewBrevoAdapter - config-dən gələn API key və sender email ilə adapter yaradır.
func NewBrevoAdapter(apiKey, senderEmail string) *BrevoAdapter {
	return &BrevoAdapter{
		apiKey:      apiKey,
		senderEmail: senderEmail,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendPasswordResetEmail - Brevo-nun /v3/smtp/email endpointinə POST edir.
// Bu metod auth.EmailService interfeysini qarşılayır.
func (b *BrevoAdapter) SendPasswordResetEmail(to string, resetURL string) error {
	// Brevo payload strukturu
	payload := map[string]interface{}{
		"sender": map[string]string{
			"name":  "Booking Support",
			"email": b.senderEmail,
		},
		"to": []map[string]string{
			{"email": to},
		},
		"subject": "Şifrəni Yenilə",
		"htmlContent": fmt.Sprintf(`
            <html>
              <body style="font-family: Arial, sans-serif;">
                <div style="padding: 20px; border: 1px solid #ddd; border-radius: 5px;">
                  <h3>Şifrəni Yenilə</h3>
                  <p><a href="%s" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none;">Şifrəni Yenilə</a></p>
                  <p style="font-size: 12px; color: #666;">Link 24 saat aktivdir.</p>
                </div>
              </body>
            </html>
        `, resetURL),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("brevo: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("brevo: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Brevo API key header
	req.Header.Set("api-key", b.apiKey)
	req.Header.Set("accept", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("brevo: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("brevo: unexpected status code %d", resp.StatusCode)
	}

	return nil
}
