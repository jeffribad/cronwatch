package notify_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/cronwatch/internal/notify"
)

func TestPagerDutyNotifier_Send_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := notify.NewPagerDutyNotifier("test-integration-key")
	n.SetPagerDutyEndpoint(ts.URL)

	if err := n.Send("job backup-db failed"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPagerDutyNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	n := notify.NewPagerDutyNotifier("bad-key")
	n.SetPagerDutyEndpoint(ts.URL)

	err := n.Send("some alert")
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestPagerDutyNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewPagerDutyNotifier("key")
	n.SetPagerDutyEndpoint("http://127.0.0.1:0")

	err := n.Send("unreachable")
	if err == nil {
		t.Fatal("expected error for bad URL, got nil")
	}
}

func TestPagerDutyNotifier_Send_PayloadContents(t *testing.T) {
	var received []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		received = buf[:n]
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	notifier := notify.NewPagerDutyNotifier("my-key")
	notifier.SetPagerDutyEndpoint(ts.URL)

	msg := "cron job nightly-report missed"
	if err := notifier.Send(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := string(received)
	for _, want := range []string{"my-key", "trigger", "cronwatch", "error", msg} {
		if !containsStr(body, want) {
			t.Errorf("payload missing %q; body: %s", want, body)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
