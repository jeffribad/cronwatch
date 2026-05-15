package notify_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestSlackNotifier_Send_Success(t *testing.T) {
	var received string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected content-type: %s", ct)
		}
		received = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewSlackNotifier(ts.URL)
	if err := n.Send("Test Subject", "Test body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = received
}

func TestSlackNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notify.NewSlackNotifier(ts.URL)
	err := n.Send("fail", "body")
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestSlackNotifier_Send_BadURL(t *testing.T) {
	n := notify.NewSlackNotifier("http://127.0.0.1:0/no-server")
	err := n.Send("fail", "body")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}
