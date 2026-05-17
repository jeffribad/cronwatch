package notify_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/cronwatch/internal/notify"
)

func TestRocketChatNotifier_Send_Success(t *testing.T) {
	var called bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewRocketChatNotifier(ts.URL, "#alerts", "cronwatch")
	notify.SetRocketChatHTTPClient(n, ts.Client())

	if err := n.Send("Job failed", "backup-db did not run"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected server to be called")
	}
}

func TestRocketChatNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notify.NewRocketChatNotifier(ts.URL, "", "")
	notify.SetRocketChatHTTPClient(n, ts.Client())

	err := n.Send("Job failed", "details")
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status 500 in error, got: %v", err)
	}
}

func TestRocketChatNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewRocketChatNotifier("http://127.0.0.1:0/no-server", "", "")
	err := n.Send("fail", "body")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestRocketChatNotifier_Send_PayloadContents(t *testing.T) {
	type payload struct {
		Text     string `json:"text"`
		Channel  string `json:"channel"`
		Username string `json:"username"`
	}

	var got payload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewRocketChatNotifier(ts.URL, "#ops", "bot")
	notify.SetRocketChatHTTPClient(n, ts.Client())
	_ = n.Send("Alert Subject", "Alert body text")

	if !strings.Contains(got.Text, "Alert Subject") {
		t.Errorf("expected subject in text, got: %s", got.Text)
	}
	if !strings.Contains(got.Text, "Alert body text") {
		t.Errorf("expected body in text, got: %s", got.Text)
	}
	if got.Channel != "#ops" {
		t.Errorf("expected channel #ops, got: %s", got.Channel)
	}
	if got.Username != "bot" {
		t.Errorf("expected username bot, got: %s", got.Username)
	}
}
