package notify

import "net/smtp"

// OverrideDial replaces the SMTP dial function, for testing only.
func (e *EmailNotifier) OverrideDial(fn func(addr string, a smtp.Auth, from string, to []string, msg []byte) error) {
	e.dial = fn
}
