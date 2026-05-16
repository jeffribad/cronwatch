package notify_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/subtlepseudonym/cronwatch/internal/notify"
)

// TestMSTeamsAdaptiveNotifier_SingleRequest ensures exactly one HTTP request is made per Send call.
func TestMSTeamsAdaptiveNotifier_SingleRequest(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewMSTeamsAdaptiveNotifier(server.URL)
	if err := n.Send("subject", "body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt32(&count); got != 1 {
		t.Errorf("expected 1 request, got %d", got)
	}
}

// TestMSTeamsAdaptiveNotifier_202Accepted treats 202 as a success.
func TestMSTeamsAdaptiveNotifier_202Accepted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	n := notify.NewMSTeamsAdaptiveNotifier(server.URL)
	if err := n.Send("subject", "body"); err != nil {
		t.Fatalf("expected 202 to be treated as success, got error: %v", err)
	}
}
