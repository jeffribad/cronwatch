package notify_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/cronwatch/internal/notify"
)

func TestVictorOpsNotifier_Send_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewVictorOpsNotifier("apikey123", "routing456")
	notify.SetVictorOpsBaseURL(n, ts.URL)

	if err := n.Send("job-failed", "cron job did not run"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestVictorOpsNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notify.NewVictorOpsNotifier("apikey123", "routing456")
	notify.SetVictorOpsBaseURL(n, ts.URL)

	err := n.Send("job-failed", "cron job did not run")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status 500 in error, got: %v", err)
	}
}

func TestVictorOpsNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewVictorOpsNotifier("key", "route")
	notify.SetVictorOpsBaseURL(n, "http://127.0.0.1:0")

	err := n.Send("subject", "body")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestVictorOpsNotifier_Send_PayloadContents(t *testing.T) {
	var captured []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		captured, err = io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewVictorOpsNotifier("k", "r")
	notify.SetVictorOpsBaseURL(n, ts.URL)

	subject := "backup-job"
	body := "missed scheduled run"
	if err := n.Send(subject, body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(captured, &payload); err != nil {
		t.Fatalf("invalid JSON payload: %v", err)
	}
	if payload["entity_id"] != subject {
		t.Errorf("expected entity_id %q, got %v", subject, payload["entity_id"])
	}
	if payload["state_message"] != body {
		t.Errorf("expected state_message %q, got %v", body, payload["state_message"])
	}
	if payload["message_type"] != "CRITICAL" {
		t.Errorf("expected message_type CRITICAL, got %v", payload["message_type"])
	}
}

// TestVictorOpsNotifier_Send_ContentTypeHeader verifies that the notifier sets
// the correct Content-Type header on outgoing requests.
func TestVictorOpsNotifier_Send_ContentTypeHeader(t *testing.T) {
	var contentType string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewVictorOpsNotifier("k", "r")
	notify.SetVictorOpsBaseURL(n, ts.URL)

	if err := n.Send("job", "body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}
}

func TestVictorOpsNotifier_DefaultBaseURL(t *testing.T) {
	n := notify.NewVictorOpsNotifier("k", "r")
	// Just ensure construction succeeds and Send returns an error (no real server).
	err := n.Send("s", "b")
	if err == nil {
		t.Log("unexpected success against real VictorOps endpoint")
	}
}
