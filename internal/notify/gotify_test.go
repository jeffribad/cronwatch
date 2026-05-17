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

func TestGotifyNotifier_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != "test-token" {
			t.Errorf("expected token 'test-token', got %q", r.URL.Query().Get("token"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewGotifyNotifier(server.URL, "test-token", 5)
	notify.SetGotifyHTTPClient(n, server.Client())

	if err := n.Send("Test Subject", "Test Body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGotifyNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	n := notify.NewGotifyNotifier(server.URL, "bad-token", 5)
	notify.SetGotifyHTTPClient(n, server.Client())

	err := n.Send("Subject", "Body")
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 in error, got: %v", err)
	}
}

func TestGotifyNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewGotifyNotifier("http://127.0.0.1:0", "token", 5)
	err := n.Send("Subject", "Body")
	if err == nil {
		t.Fatal("expected error for bad URL")
	}
}

func TestGotifyNotifier_Send_PayloadContents(t *testing.T) {
	type gotifyPayload struct {
		Title    string `json:"title"`
		Message  string `json:"message"`
		Priority int    `json:"priority"`
	}

	var received gotifyPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewGotifyNotifier(server.URL, "token", 8)
	notify.SetGotifyHTTPClient(n, server.Client())

	_ = n.Send("Alert: job failed", "cron-backup missed its window")

	if received.Title != "Alert: job failed" {
		t.Errorf("expected title 'Alert: job failed', got %q", received.Title)
	}
	if received.Message != "cron-backup missed its window" {
		t.Errorf("unexpected message: %q", received.Message)
	}
	if received.Priority != 8 {
		t.Errorf("expected priority 8, got %d", received.Priority)
	}
}

func TestGotifyNotifier_DefaultPriority(t *testing.T) {
	type gotifyPayload struct {
		Priority int `json:"priority"`
	}

	var received gotifyPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// priority 0 should default to 5
	n := notify.NewGotifyNotifier(server.URL, "token", 0)
	notify.SetGotifyHTTPClient(n, server.Client())
	_ = n.Send("s", "b")

	if received.Priority != 5 {
		t.Errorf("expected default priority 5, got %d", received.Priority)
	}
}
