package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/api"
	"github.com/user/cronwatch/internal/history"
)

func tempHistoryStore(t *testing.T) *history.Store {
	t.Helper()
	f, err := os.CreateTemp("", "api-history-*.json")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	store, err := history.NewStore(f.Name())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return store
}

func TestHistory_Empty(t *testing.T) {
	h := api.NewHandler(nil, tempHistoryStore(t))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("expected empty map, got %v", body)
	}
}

func TestHistory_WithEntries(t *testing.T) {
	store := tempHistoryStore(t)
	_ = store.Record("backup", history.Entry{
		JobName:   "backup",
		Success:   true,
		Timestamp: time.Now(),
	})

	h := api.NewHandler(nil, store)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string][]map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body["backup"]) != 1 {
		t.Errorf("expected 1 backup entry, got %d", len(body["backup"]))
	}
}

func TestHistory_MethodNotAllowed(t *testing.T) {
	h := api.NewHandler(nil, tempHistoryStore(t))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/history", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
