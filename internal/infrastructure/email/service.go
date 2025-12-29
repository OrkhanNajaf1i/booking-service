package email

import (
	"fmt"
	"log"
)

type DummyEmailService struct{}

func NewDummyEmailService() *DummyEmailService {
	return &DummyEmailService{}
}
func (s *DummyEmailService) SendWelcomeEmail(to string, name string) error {
	log.Printf("[EMAIL MOCK] ✉️  To: %s | Subject: Welcome %s! | Body: Thank you for registering.", to, name)
	return nil
}

// SendPasswordResetEmail - Parol sıfırlama linki göndərir
func (s *DummyEmailService) SendPasswordResetEmail(to string, token string) error {
	resetLink := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)

	log.Printf("[EMAIL MOCK] ✉️  To: %s", to)
	log.Printf("[EMAIL MOCK] 📧 Subject: Reset Your Password")
	log.Printf("[EMAIL MOCK] 🔗 Link: %s", resetLink)
	log.Printf("[EMAIL MOCK] ⏰ This link expires in 24 hours")

	return nil
}

// Gələcək funksiyalar (opsional):
// - SendVerificationEmail(to, verificationCode string) error
// - SendBookingConfirmation(to string, bookingDetails map[string]interface{}) error
