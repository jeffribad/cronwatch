package notify_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestTelegramNotifier_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewTelegramNotifier("test-token", "chat-123")
	notify.SetTelegramBaseURL(n, server.URL)
	notify.SetTelegramHTTPClient(n, server.Client())

	if err := n.Send("Alert", "job failed"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTelegramNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	n := notify.NewTelegramNotifier("bad-token", "chat-123")
	notify.SetTelegramBaseURL(n, server.URL)
	notify.SetTelegramHTTPClient(n, server.Client())

	err := n.Send("Alert", "job failed")
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected status 401 in error, got: %v", err)
	}
}

func TestTelegramNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewTelegramNotifier("token", "chat-123")
	notify.SetTelegramBaseURL(n, "http://127.0.0.1:0")

	err := n.Send("Alert", "body")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestTelegramNotifier_Send_PayloadContents(t *testing.T) {
	var captured []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewTelegramNotifier("tok", "999")
	notify.SetTelegramBaseURL(n, server.URL)
	notify.SetTelegramHTTPClient(n, server.Client())

	_ = n.Send("My Subject", "My Body")

	var payload map[string]interface{}
	if err := json.Unmarshal(captured, &payload); err != nil {
		t.Fatalf("invalid JSON payload: %v", err)
	}
	if payload["chat_id"] != "999" {
		t.Errorf("expected chat_id=999, got %v", payload["chat_id"])
	}
	text, _ := payload["text"].(string)
	if !strings.Contains(text, "My Subject") || !strings.Contains(text, "My Body") {
		t.Errorf("expected text to contain subject and body, got: %s", text)
	}
	if payload["parse_mode"] != "Markdown" {
		t.Errorf("expected parse_mode=Markdown, got %v", payload["parse_mode"])
	}
}
