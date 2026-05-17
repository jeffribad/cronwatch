package notify_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/user/cronwatch/internal/notify"
)

// TestMattermostNotifier_SingleRequest verifies exactly one HTTP request is made per Send call.
func TestMattermostNotifier_SingleRequest(t *testing.T) {
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notify.NewMattermostNotifier(server.URL, "", "")
	notify.SetMattermostHTTPClient(n, server.Client())

	if err := n.Send("title", "body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("expected 1 request, got %d", count)
	}
}

// TestMattermostNotifier_201Accepted verifies 201 is also treated as success.
func TestMattermostNotifier_201Accepted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	n := notify.NewMattermostNotifier(server.URL, "", "")
	notify.SetMattermostHTTPClient(n, server.Client())

	if err := n.Send("title", "body"); err != nil {
		t.Errorf("expected 201 to be accepted, got error: %v", err)
	}
}
