package email

import (
	"fmt"
	"log"
	"net/smtp"
)

type SMTPService struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewSMTPService(host string, port int, username, password, from string) *SMTPService {
	log.Printf("[SMTP DEBUG] Init -> Host: %s:%d | User: %s | Sender: %s", host, port, username, from)
	return &SMTPService{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func (s *SMTPService) SendPasswordResetEmail(to string, resetLink string) error {
	senderHeader := fmt.Sprintf("Booking Support <%s>", s.From)

	subject := "Şifrə Yeniləmə Tələbi"
	headers := map[string]string{
		"From":         senderHeader,
		"To":           to,
		"Subject":      subject,
		"Reply-To":     s.From,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}

	headerStr := ""
	for k, v := range headers {
		headerStr += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	headerStr += "\r\n"

	body := fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif;">
				<div style="padding: 20px; border: 1px solid #ddd; border-radius: 5px;">
					<h3>Şifrəni Yenilə</h3>
					<p><a href="%s" style="background-color: #007bff; color: white; padding: 10px 20px; text-decoration: none;">Şifrəni Yenilə</a></p>
					<p style="font-size: 12px; color: #666;">Link 24 saat aktivdir.</p>
				</div>
			</body>
		</html>
	`, resetLink)

	msg := []byte(headerStr + body)

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	log.Printf("[SMTP DEBUG] Sending to %s...", to)

	err := smtp.SendMail(addr, auth, s.From, []string{to}, msg)
	if err != nil {
		log.Printf("❌ [SMTP ERROR] %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	log.Printf("[DEBUG-EMAIL] Dialing %s...", addr)
	conn, err := smtp.Dial(addr)
	if err != nil {
		log.Printf("[DEBUG-EMAIL] ❌ Bağlantı qurula bilmədi: %v", err)
		return err
	}
	defer conn.Close()

	log.Printf("[DEBUG-EMAIL] 🔍 EHLO/HELO yoxlanılır...")
	if err = conn.Hello("localhost"); err != nil {
		log.Printf("[DEBUG-EMAIL] ❌ Hello xətası: %v", err)
		return err
	}

	log.Printf("[DEBUG-EMAIL] 🔐 Autentifikasiya yoxlanılır...")
	if err = conn.Auth(auth); err != nil {
		log.Printf("[DEBUG-EMAIL] ❌ Şifrə/User xətası: %v", err)
		return err
	}
	log.Printf("✅ [SMTP SUCCESS] Mail göndərildi!")
	return nil
}

func (s *SMTPService) SendWelcomeEmail(to string, name string) error {
	return nil
}
