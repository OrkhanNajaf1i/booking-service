package auth

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/smtp"
)

type SMTPEmailService struct {
	smtpHost       string
	smtpPort       string
	senderEmail    string
	senderPassword string
	appName        string
}

func NewSMTPEmailService(host, port, email, password, appName string) *SMTPEmailService {
	return &SMTPEmailService{
		smtpHost:       host,
		smtpPort:       port,
		senderEmail:    email,
		senderPassword: password,
		appName:        appName,
	}
}

func (s *SMTPEmailService) SendPasswordResetEmail(email, resetURL string) error {
	subject := fmt.Sprintf("Parol sıfırlama %s", s.appName)
	plainBody := fmt.Sprintf("Parol Sıfırlama\n\nParolunuzu sıfırlamaq üçün: %s\n\nLink 24 saat etibarlıdır.\n%s", resetURL, s.appName)

	htmlBody := fmt.Sprintf(`
							<html><body style="font-family:Arial,sans-serif;">
							<h2>Parol Sıfırlama</h2>
							<p>Parolunuzu sıfırlamaq üçün linki klikləyin:</p>
							<a href="%s" style="background:#007bff;color:white;padding:10px 20px;text-decoration:none;display:inline-block;">Parolumu Sıfırla</a>
							<p><strong>Link:</strong> <code>%s</code> (24 saat etibarlıdır)</p>
							</body></html>
							`, resetURL, resetURL)
	var msg bytes.Buffer
	writer := multipart.NewWriter(&msg)
	msg.WriteString(fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%s\r\n\r\n--%s\r\n",
		email, subject, writer.Boundary(), writer.Boundary()))
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	msg.WriteString(plainBody)
	msg.WriteString(fmt.Sprintf("\r\n--%s\r\n", writer.Boundary()))

	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString(fmt.Sprintf("\r\n--%s--\r\n", writer.Boundary()))

	auth := smtp.PlainAuth("", s.senderEmail, s.senderPassword, s.smtpHost)
	return smtp.SendMail(s.smtpHost+":"+s.smtpPort, auth, s.senderEmail, []string{email}, msg.Bytes())
}
