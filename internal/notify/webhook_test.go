package notify_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestWebhookNotifier_Send_Success(t *testing.T) {
	var received map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewWebhookNotifier(ts.URL)
	if err := n.Send("backup-job", "failure", "exit code 1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["job_name"] != "backup-job" {
		t.Errorf("expected job_name backup-job, got %v", received["job_name"])
	}
	if received["status"] != "failure" {
		t.Errorf("expected status failure, got %v", received["status"])
	}
	if received["message"] != "exit code 1" {
		t.Errorf("expected message 'exit code 1', got %v", received["message"])
	}
	if received["timestamp"] == nil {
		t.Error("expected timestamp to be set")
	}
}

func TestWebhookNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notify.NewWebhookNotifier(ts.URL)
	err := n.Send("job", "failure", "msg")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestWebhookNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewWebhookNotifier("http://127.0.0.1:0/nowhere")
	notify.SetWebhookHTTPClient(n, &http.Client{})
	err := n.Send("job", "failure", "msg")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestWebhookNotifier_Send_PayloadContents(t *testing.T) {
	var received map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := notify.NewWebhookNotifier(ts.URL)
	if err := n.Send("nightly-sync", "missed", "no run in 2h"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["status"] != "missed" {
		t.Errorf("expected status missed, got %v", received["status"])
	}
	if received["message"] != "no run in 2h" {
		t.Errorf("expected message 'no run in 2h', got %v", received["message"])
	}
}
