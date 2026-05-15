package notify_test

import (
	"errors"
	"net/smtp"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestEmailNotifier_Send_Success(t *testing.T) {
	var capturedAddr string
	var capturedFrom string
	var capturedTo []string

	cfg := notify.SMTPConfig{
		Host: "localhost", Port: 25,
		From: "alerts@example.com",
		To:   []string{"ops@example.com"},
	}
	n := notify.NewEmailNotifier(cfg)
	n.OverrideDial(func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		capturedAddr = addr
		capturedFrom = from
		capturedTo = to
		return nil
	})

	if err := n.Send("cron failed", "job backup missed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedAddr != "localhost:25" {
		t.Errorf("unexpected addr: %s", capturedAddr)
	}
	if capturedFrom != "alerts@example.com" {
		t.Errorf("unexpected from: %s", capturedFrom)
	}
	if len(capturedTo) == 0 || capturedTo[0] != "ops@example.com" {
		t.Errorf("unexpected to: %v", capturedTo)
	}
}

func TestEmailNotifier_Send_NoRecipients(t *testing.T) {
	cfg := notify.SMTPConfig{Host: "localhost", Port: 25, From: "a@b.com"}
	n := notify.NewEmailNotifier(cfg)
	if err := n.Send("x", "y"); err == nil {
		t.Fatal("expected error for empty recipients")
	}
}

func TestEmailNotifier_Send_DialError(t *testing.T) {
	cfg := notify.SMTPConfig{
		Host: "localhost", Port: 25,
		From: "a@b.com",
		To:   []string{"x@y.com"},
	}
	n := notify.NewEmailNotifier(cfg)
	n.OverrideDial(func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return errors.New("connection refused")
	})
	if err := n.Send("x", "y"); err == nil {
		t.Fatal("expected dial error")
	}
}
