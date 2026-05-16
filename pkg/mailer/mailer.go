package mailer

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// Mailer defines the email-sending contract.
type Mailer interface {
	SendVerificationEmail(to, token string) error
	SendPasswordResetEmail(to, token string) error
}

// SMTPMailer sends real emails via SMTP.
type SMTPMailer struct {
	host string
	port int
	user string
	pass string
	from string
}

func NewSMTPMailer(host string, port int, user, pass, from string) *SMTPMailer {
	return &SMTPMailer{host: host, port: port, user: user, pass: pass, from: from}
}

func (m *SMTPMailer) SendVerificationEmail(to, token string) error {
	link := fmt.Sprintf("https://your-domain.com/verify-email?token=%s", token)
	body := fmt.Sprintf("Click the link below to verify your email:\n\n%s\n\nThis link expires in 24 hours.", link)
	return m.send(to, "Verify Your Email — Kamus Banjar", body)
}

func (m *SMTPMailer) SendPasswordResetEmail(to, token string) error {
	link := fmt.Sprintf("https://your-domain.com/reset-password?token=%s", token)
	body := fmt.Sprintf("Click the link below to reset your password:\n\n%s\n\nThis link expires in 1 hour.", link)
	return m.send(to, "Password Reset — Kamus Banjar", body)
}

func (m *SMTPMailer) send(to, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	d := gomail.NewDialer(m.host, m.port, m.user, m.pass)
	return d.DialAndSend(msg)
}

// MockMailer captures sent emails for testing.
type MockMailer struct {
	VerificationCalls []SentEmail
	ResetCalls        []SentEmail
}

type SentEmail struct {
	To    string
	Token string
}

func (m *MockMailer) SendVerificationEmail(to, token string) error {
	m.VerificationCalls = append(m.VerificationCalls, SentEmail{To: to, Token: token})
	return nil
}

func (m *MockMailer) SendPasswordResetEmail(to, token string) error {
	m.ResetCalls = append(m.ResetCalls, SentEmail{To: to, Token: token})
	return nil
}
