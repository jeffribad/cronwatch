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

func TestMattermostNotifier_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewMattermostNotifier(server.URL, "#alerts", "cronwatch")
	notify.SetMattermostHTTPClient(n, server.Client())

	if err := n.Send("Test Subject", "Test body"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMattermostNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := notify.NewMattermostNotifier(server.URL, "", "")
	notify.SetMattermostHTTPClient(n, server.Client())

	if err := n.Send("Subject", "Body"); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestMattermostNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewMattermostNotifier("http://127.0.0.1:0/no-such-server", "", "")
	if err := n.Send("Subject", "Body"); err == nil {
		t.Fatal("expected error for bad URL")
	}
}

func TestMattermostNotifier_Send_PayloadContents(t *testing.T) {
	var captured []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewMattermostNotifier(server.URL, "#ops", "bot")
	notify.SetMattermostHTTPClient(n, server.Client())

	_ = n.Send("Alert Title", "Something went wrong")

	var payload map[string]string
	if err := json.Unmarshal(captured, &payload); err != nil {
		t.Fatalf("invalid JSON payload: %v", err)
	}
	if !strings.Contains(payload["text"], "Alert Title") {
		t.Errorf("expected text to contain subject, got: %s", payload["text"])
	}
	if !strings.Contains(payload["text"], "Something went wrong") {
		t.Errorf("expected text to contain body, got: %s", payload["text"])
	}
	if payload["channel"] != "#ops" {
		t.Errorf("expected channel #ops, got: %s", payload["channel"])
	}
	if payload["username"] != "bot" {
		t.Errorf("expected username bot, got: %s", payload["username"])
	}
}

func TestMattermostNotifier_Send_ContentTypeHeader(t *testing.T) {
	var contentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewMattermostNotifier(server.URL, "", "")
	notify.SetMattermostHTTPClient(n, server.Client())
	_ = n.Send("Subject", "Body")

	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got: %s", contentType)
	}
}
