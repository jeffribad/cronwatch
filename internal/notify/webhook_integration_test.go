package notify_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

// TestWebhookNotifier_RetriesAreNotPerformed ensures the notifier does not
// silently retry on failure — the caller (dispatcher) owns retry logic.
func TestWebhookNotifier_RetriesAreNotPerformed(t *testing.T) {
	var callCount int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	n := notify.NewWebhookNotifier(ts.URL)
	_ = n.Send("job", "failure", "something went wrong")

	if got := atomic.LoadInt32(&callCount); got != 1 {
		t.Errorf("expected exactly 1 call, got %d", got)
	}
}

// TestWebhookNotifier_201Accepted treats 201 as success.
func TestWebhookNotifier_201Accepted(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	n := notify.NewWebhookNotifier(ts.URL)
	if err := n.Send("job", "success", "all good"); err != nil {
		t.Fatalf("expected no error for 201, got: %v", err)
	}
}
