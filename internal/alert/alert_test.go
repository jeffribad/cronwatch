package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/alert"
)

func TestWebhookNotifier_Send_Success(t *testing.T) {
	var received map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := alert.NewWebhookNotifier(server.URL, 5*time.Second)
	if err := n.Send("job failed", "backup-db exited with code 1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if received["subject"] != "job failed" {
		t.Errorf("subject mismatch: %q", received["subject"])
	}
	if received["body"] != "backup-db exited with code 1" {
		t.Errorf("body mismatch: %q", received["body"])
	}
	if received["timestamp"] == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestWebhookNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := alert.NewWebhookNotifier(server.URL, 5*time.Second)
	err := n.Send("test", "body")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status 500 in error, got: %v", err)
	}
}

func TestWebhookNotifier_Send_BadURL(t *testing.T) {
	n := alert.NewWebhookNotifier("http://127.0.0.1:0/no-server", 1*time.Second)
	if err := n.Send("x", "y"); err == nil {
		t.Fatal("expected connection error")
	}
}

type mockWriter struct{ buf strings.Builder }

func (m *mockWriter) WriteString(s string) (int, error) { return m.buf.WriteString(s) }

func TestLogNotifier_Send(t *testing.T) {
	w := &mockWriter{}
	n := &alert.LogNotifier{Writer: w}
	if err := n.Send("missed run", "nightly-report"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := w.buf.String()
	if !strings.Contains(out, "[ALERT]") {
		t.Errorf("expected [ALERT] in output: %q", out)
	}
	if !strings.Contains(out, "missed run") {
		t.Errorf("expected subject in output: %q", out)
	}
}
