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

func TestTeamsNotifier_Send_Success(t *testing.T) {
	var received []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewTeamsNotifier(ts.URL)
	if err := n.Send("Job failed", "backup-db did not run"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(received, &payload); err != nil {
		t.Fatalf("invalid JSON payload: %v", err)
	}
	if payload["@type"] != "MessageCard" {
		t.Errorf("expected @type MessageCard, got %v", payload["@type"])
	}
}

func TestTeamsNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	n := notify.NewTeamsNotifier(ts.URL)
	err := n.Send("Job failed", "some detail")
	if err == nil {
		t.Fatal("expected error on non-200 status")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("expected status 400 in error, got: %v", err)
	}
}

func TestTeamsNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewTeamsNotifier("http://127.0.0.1:0/no-listener")
	err := n.Send("Job failed", "detail")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestTeamsNotifier_Send_PayloadContents(t *testing.T) {
	var body []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewTeamsNotifier(ts.URL)
	_ = n.Send("Alert subject", "Alert body text")

	if !strings.Contains(string(body), "Alert subject") {
		t.Errorf("payload missing subject; got: %s", body)
	}
	if !strings.Contains(string(body), "Alert body text") {
		t.Errorf("payload missing body; got: %s", body)
	}
	if !strings.Contains(string(body), "FF0000") {
		t.Errorf("payload missing themeColor; got: %s", body)
	}
}
