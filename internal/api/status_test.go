package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/history"
)

func tempStatusStore(t *testing.T) *history.Store {
	t.Helper()
	store, err := history.NewStore(t.TempDir() + "/status_history.json")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	return store
}

func statusConfig() *config.Config {
	return &config.Config{
		Jobs: []config.Job{
			{Name: "backup", Schedule: "0 2 * * *"},
			{Name: "cleanup", Schedule: "0 3 * * *"},
		},
	}
}

func TestStatus_AllUnknown(t *testing.T) {
	store := tempStatusStore(t)
	h := NewStatusHandler(store, statusConfig())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var statuses []JobStatus
	if err := json.NewDecoder(rr.Body).Decode(&statuses); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	for _, s := range statuses {
		if s.Status != "unknown" {
			t.Errorf("expected unknown, got %s for %s", s.Status, s.Name)
		}
	}
}

func TestStatus_WithSuccessRecord(t *testing.T) {
	store := tempStatusStore(t)
	cfg := statusConfig()

	store.Record(history.Entry{Job: "backup", Time: time.Now(), Success: true})

	h := NewStatusHandler(store, cfg)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	h.ServeHTTP(rr, req)

	var statuses []JobStatus
	json.NewDecoder(rr.Body).Decode(&statuses)

	for _, s := range statuses {
		if s.Name == "backup" && s.Status != "ok" {
			t.Errorf("expected ok for backup, got %s", s.Status)
		}
	}
}

func TestStatus_MethodNotAllowed(t *testing.T) {
	store := tempStatusStore(t)
	h := NewStatusHandler(store, statusConfig())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/status", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
