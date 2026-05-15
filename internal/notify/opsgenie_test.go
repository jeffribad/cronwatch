package notify_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/cronwatch/internal/notify"
)

func TestOpsGenieNotifier_Send_Success(t *testing.T) {
	var captured map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := notify.NewOpsGenieNotifier("test-key", ts.URL)
	if err := n.Send("job failed", "details here"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured["message"] != "job failed" {
		t.Errorf("expected message 'job failed', got %v", captured["message"])
	}
	if captured["description"] != "details here" {
		t.Errorf("expected description 'details here', got %v", captured["description"])
	}
}

func TestOpsGenieNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	n := notify.NewOpsGenieNotifier("bad-key", ts.URL)
	err := n.Send("subject", "body")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 in error, got: %v", err)
	}
}

func TestOpsGenieNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewOpsGenieNotifier("key", "http://127.0.0.1:0")
	err := n.Send("subject", "body")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestOpsGenieNotifier_Send_PayloadContents(t *testing.T) {
	var payload map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &payload)
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	n := notify.NewOpsGenieNotifier("key", ts.URL)
	_ = n.Send("alert", "desc")

	if payload["priority"] != "P2" {
		t.Errorf("expected priority P2, got %v", payload["priority"])
	}
	tags, ok := payload["tags"].([]interface{})
	if !ok || len(tags) == 0 || tags[0] != "cronwatch" {
		t.Errorf("expected tags to contain 'cronwatch', got %v", payload["tags"])
	}
}

func TestOpsGenieNotifier_DefaultBaseURL(t *testing.T) {
	n := notify.NewOpsGenieNotifier("key", "")
	if n.OpsGenieBaseURL() != "https://api.opsgenie.com/v2/alerts" {
		t.Errorf("unexpected default base URL: %s", n.OpsGenieBaseURL())
	}
}
