package api_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/user/cronwatch/internal/alert"
	"github.com/user/cronwatch/internal/api"
	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/history"
	"github.com/user/cronwatch/internal/monitor"
)

func setup(t *testing.T) (*api.Handler, *history.Store) {
	t.Helper()
	cfg := &config.Config{
		CheckIntervalSeconds: 60,
		Jobs: []config.Job{
			{Name: "backup", MaxAgeSeconds: 3600},
		},
	}
	store, err := history.NewStore(t.TempDir() + "/h.json")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	logger := log.New(os.Stderr, "", 0)
	dispatcher := alert.NewDispatcher(nil, logger)
	mon := monitor.New(cfg, store, dispatcher, logger)
	return api.NewHandler(mon, store, logger), store
}

func TestHealthz(t *testing.T) {
	h, _ := setup(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReport_Success(t *testing.T) {
	h, _ := setup(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body, _ := json.Marshal(map[string]interface{}{"job": "backup", "success": true})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/report", bytes.NewReader(body)))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}

func TestReport_MissingJob(t *testing.T) {
	h, _ := setup(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body, _ := json.Marshal(map[string]interface{}{"success": true})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/report", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestReport_MethodNotAllowed(t *testing.T) {
	h, _ := setup(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/report", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestListJobs_Empty(t *testing.T) {
	h, _ := setup(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/jobs", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
