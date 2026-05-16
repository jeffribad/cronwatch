package notify_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/subtlepseudonym/cronwatch/internal/notify"
)

func TestMSTeamsAdaptiveNotifier_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewMSTeamsAdaptiveNotifier(server.URL)
	if err := n.Send("Alert", "Job failed"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMSTeamsAdaptiveNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := notify.NewMSTeamsAdaptiveNotifier(server.URL)
	if err := n.Send("Alert", "Job failed"); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestMSTeamsAdaptiveNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewMSTeamsAdaptiveNotifier("http://127.0.0.1:0/no-server")
	if err := n.Send("Alert", "Job failed"); err == nil {
		t.Fatal("expected error for bad URL")
	}
}

func TestMSTeamsAdaptiveNotifier_Send_PayloadContents(t *testing.T) {
	var captured []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewMSTeamsAdaptiveNotifier(server.URL)
	if err := n.Send("Test Subject", "Test Body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(captured, &payload); err != nil {
		t.Fatalf("invalid JSON payload: %v", err)
	}

	raw, _ := json.Marshal(payload)
	s := string(raw)
	if !strings.Contains(s, "Test Subject") {
		t.Errorf("payload missing subject: %s", s)
	}
	if !strings.Contains(s, "Test Body") {
		t.Errorf("payload missing body: %s", s)
	}
	if !strings.Contains(s, "AdaptiveCard") {
		t.Errorf("payload missing AdaptiveCard type: %s", s)
	}
}

func TestMSTeamsAdaptiveNotifier_Send_ContentTypeHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewMSTeamsAdaptiveNotifier(server.URL)
	if err := n.Send("Subject", "Body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
