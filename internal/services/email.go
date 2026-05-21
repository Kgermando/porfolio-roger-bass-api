package services

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// EmailService handles outgoing emails via SMTP
type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
	adminTo  string
}

// NewEmailService creates an EmailService from environment variables.
// Returns nil when SMTP credentials are absent.
func NewEmailService() *EmailService {
	host := os.Getenv("EMAIL_HOST")
	port := os.Getenv("EMAIL_PORT")
	username := os.Getenv("EMAIL_USERNAME")
	password := os.Getenv("EMAIL_PASSWORD")
	from := os.Getenv("EMAIL_FROM")
	adminTo := os.Getenv("EMAIL_ADMIN_TO")

	if host == "" || port == "" || username == "" || password == "" {
		return nil
	}

	return &EmailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		adminTo:  adminTo,
	}
}

// SendContactNotification notifies the admin about a new contact form submission.
func (es *EmailService) SendContactNotification(name, senderEmail, subject, message, phone string) error {
	if es == nil || es.adminTo == "" {
		return nil // silently skip when not configured
	}

	displaySubject := subject
	if displaySubject == "" {
		displaySubject = "Nouveau message de contact"
	}

	body := buildContactEmailHTML(name, senderEmail, phone, displaySubject, message)

	return es.sendEmail(es.adminTo, "[Roger Bass Portfolio] "+displaySubject, body)
}

// sendEmail sends an HTML email via SMTP
func (es *EmailService) sendEmail(to, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", es.username, es.password, es.host)

	header := strings.Join([]string{
		"From: " + es.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
	}, "\r\n")

	msg := []byte(header + "\r\n" + htmlBody)

	addr := fmt.Sprintf("%s:%s", es.host, es.port)
	return smtp.SendMail(addr, auth, es.from, []string{to}, msg)
}

func buildContactEmailHTML(name, email, phone, subject, message string) string {
	phoneRow := ""
	if phone != "" {
		phoneRow = fmt.Sprintf(`<tr><td><strong>Téléphone</strong></td><td>%s</td></tr>`, htmlEscape(phone))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="UTF-8" />
  <title>Nouveau message</title>
</head>
<body style="font-family:Arial,sans-serif;background:#f4f4f4;margin:0;padding:0;">
  <div style="max-width:600px;margin:30px auto;background:#fff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,.1);">
    <div style="background:#1a1a2e;padding:20px 30px;color:#fff;">
      <h2 style="margin:0;">Nouveau message — Roger Bass Portfolio</h2>
    </div>
    <div style="padding:30px;">
      <table style="width:100%%;border-collapse:collapse;font-size:15px;">
        <tr>
          <td style="padding:8px 0;width:140px;color:#555;"><strong>Nom</strong></td>
          <td style="padding:8px 0;">%s</td>
        </tr>
        <tr>
          <td style="padding:8px 0;color:#555;"><strong>Email</strong></td>
          <td style="padding:8px 0;"><a href="mailto:%s">%s</a></td>
        </tr>
        %s
        <tr>
          <td style="padding:8px 0;color:#555;"><strong>Sujet</strong></td>
          <td style="padding:8px 0;">%s</td>
        </tr>
      </table>
      <hr style="border:none;border-top:1px solid #eee;margin:20px 0;" />
      <p style="font-size:15px;color:#333;white-space:pre-wrap;">%s</p>
    </div>
    <div style="background:#f0f0f0;padding:12px 30px;font-size:12px;color:#888;text-align:center;">
      Ce message a été envoyé depuis le formulaire de contact de rogerbass.com
    </div>
  </div>
</body>
</html>`,
		htmlEscape(name),
		htmlEscape(email), htmlEscape(email),
		phoneRow,
		htmlEscape(subject),
		htmlEscape(message),
	)
}

// htmlEscape escapes basic HTML special chars to prevent injection in emails
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&#34;")
	return s
}
