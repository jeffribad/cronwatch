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

func TestEmailNotifier_Send_MessageContainsSubjectAndBody(t *testing.T) {
	var capturedMsg []byte

	cfg := notify.SMTPConfig{
		Host: "localhost", Port: 25,
		From: "alerts@example.com",
		To:   []string{"ops@example.com"},
	}
	n := notify.NewEmailNotifier(cfg)
	n.OverrideDial(func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		capturedMsg = msg
		return nil
	})

	subject := "cron failed"
	body := "job backup missed"
	if err := n.Send(subject, body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msgStr := string(capturedMsg)
	if !contains(msgStr, subject) {
		t.Errorf("expected message to contain subject %q, got: %s", subject, msgStr)
	}
	if !contains(msgStr, body) {
		t.Errorf("expected message to contain body %q, got: %s", body, msgStr)
	}
}

// contains is a simple helper to check substring presence.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		})())
}
