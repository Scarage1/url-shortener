package email

import (
	"fmt"
	"log"
	"net/smtp"
)

// Sender sends transactional emails via SMTP.
type Sender interface {
	Send(to, subject, body string) error
}

// SMTPSender sends emails via an SMTP server (Gmail, SendGrid, etc.).
type SMTPSender struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

func NewSMTPSender(host, port, user, password, from string) *SMTPSender {
	return &SMTPSender{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		From:     from,
	}
}

func (s *SMTPSender) Send(to, subject, body string) error {

	if s.Host == "" {
		log.Printf("SMTP not configured — would send to %s: %s", to, subject)
		return nil
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		s.From, to, subject, body,
	)

	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)
	addr := s.Host + ":" + s.Port

	return smtp.SendMail(addr, auth, s.From, []string{to}, []byte(msg))
}

// NoopSender logs emails instead of sending them. Used in development/tests.
type NoopSender struct{}

func (NoopSender) Send(to, subject, body string) error {
	log.Printf("[EMAIL] To: %s | Subject: %s | Body length: %d", to, subject, len(body))
	return nil
}
