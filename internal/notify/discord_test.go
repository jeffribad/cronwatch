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

func TestDiscordNotifier_Send_Success(t *testing.T) {
	var gotBody map[string]string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	n := notify.NewDiscordNotifier(ts.URL)
	notify.SetDiscordHTTPClient(n, ts.Client())

	if err := n.Send("Job failed", "cron job backup did not run"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(gotBody["content"], "Job failed") {
		t.Errorf("expected content to contain subject, got %q", gotBody["content"])
	}
	if !strings.Contains(gotBody["content"], "cron job backup did not run") {
		t.Errorf("expected content to contain body, got %q", gotBody["content"])
	}
}

func TestDiscordNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notify.NewDiscordNotifier(ts.URL)
	notify.SetDiscordHTTPClient(n, ts.Client())

	if err := n.Send("subject", "body"); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestDiscordNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewDiscordNotifier("http://127.0.0.1:0/no-server")
	if err := n.Send("subject", "body"); err == nil {
		t.Fatal("expected error for bad URL, got nil")
	}
}

func TestDiscordNotifier_Send_PayloadContents(t *testing.T) {
	var gotContent string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &payload)
		gotContent = payload["content"]
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	n := notify.NewDiscordNotifier(ts.URL)
	notify.SetDiscordHTTPClient(n, ts.Client())

	_ = n.Send("Alert", "details here")

	if !strings.HasPrefix(gotContent, "**Alert**") {
		t.Errorf("expected bold subject prefix, got %q", gotContent)
	}
	if !strings.Contains(gotContent, "details here") {
		t.Errorf("expected body in content, got %q", gotContent)
	}
}
