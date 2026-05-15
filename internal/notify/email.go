package notify

import (
	"fmt"
	"net/smtp"
)

// SMTPConfig holds the configuration for sending email alerts.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

// EmailNotifier sends alert messages via SMTP.
type EmailNotifier struct {
	cfg  SMTPConfig
	dial func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// NewEmailNotifier creates an EmailNotifier with the given SMTP configuration.
func NewEmailNotifier(cfg SMTPConfig) *EmailNotifier {
	return &EmailNotifier{cfg: cfg, dial: smtp.SendMail}
}

// Send delivers an email with the given subject and body to all configured recipients.
func (e *EmailNotifier) Send(subject, body string) error {
	if len(e.cfg.To) == 0 {
		return fmt.Errorf("email: no recipients configured")
	}

	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)
	var auth smtp.Auth
	if e.cfg.Username != "" {
		auth = smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.Host)
	}

	header := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n",
		e.cfg.From, e.cfg.To[0], subject)
	msg := []byte(header + body)

	if err := e.dial(addr, auth, e.cfg.From, e.cfg.To, msg); err != nil {
		return fmt.Errorf("email: send: %w", err)
	}
	return nil
}
