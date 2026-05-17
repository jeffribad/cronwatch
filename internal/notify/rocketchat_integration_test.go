package notify_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/your-org/cronwatch/internal/notify"
)

// TestRocketChatNotifier_SingleRequest verifies exactly one HTTP request is made per Send call.
func TestRocketChatNotifier_SingleRequest(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewRocketChatNotifier(ts.URL, "#general", "cronwatch")
	notify.SetRocketChatHTTPClient(n, ts.Client())

	if err := n.Send("test subject", "test body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c := atomic.LoadInt32(&count); c != 1 {
		t.Errorf("expected 1 request, got %d", c)
	}
}

// TestRocketChatNotifier_201Accepted verifies 201 is treated as success.
func TestRocketChatNotifier_201Accepted(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	n := notify.NewRocketChatNotifier(ts.URL, "", "")
	notify.SetRocketChatHTTPClient(n, ts.Client())

	if err := n.Send("subject", "body"); err != nil {
		t.Fatalf("expected 201 to be accepted, got error: %v", err)
	}
}
