package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/history"
)

func tempMetricsStore(t *testing.T) *history.Store {
	t.Helper()
	path := t.TempDir() + "/metrics_history.json"
	s, err := history.NewStore(path)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	return s
}

func TestMetrics_Empty(t *testing.T) {
	store := tempMetricsStore(t)
	h := NewMetricsHandler(store, []string{"backup"})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var summary MetricsSummary
	if err := json.NewDecoder(rr.Body).Decode(&summary); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if summary.TotalJobs != 1 {
		t.Errorf("expected 1 job, got %d", summary.TotalJobs)
	}
	if summary.Jobs[0].LastStatus != "unknown" {
		t.Errorf("expected unknown status, got %s", summary.Jobs[0].LastStatus)
	}
}

func TestMetrics_WithEntries(t *testing.T) {
	store := tempMetricsStore(t)
	now := time.Now()
	_ = store.Record("backup", history.Entry{JobName: "backup", Success: true, RunAt: now.Add(-2 * time.Minute)})
	_ = store.Record("backup", history.Entry{JobName: "backup", Success: false, RunAt: now.Add(-1 * time.Minute)})
	_ = store.Record("backup", history.Entry{JobName: "backup", Success: true, RunAt: now})

	h := NewMetricsHandler(store, []string{"backup"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rr, req)

	var summary MetricsSummary
	if err := json.NewDecoder(rr.Body).Decode(&summary); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if summary.TotalRuns != 3 {
		t.Errorf("expected 3 runs, got %d", summary.TotalRuns)
	}
	if summary.TotalFailures != 1 {
		t.Errorf("expected 1 failure, got %d", summary.TotalFailures)
	}
	expectedRate := 200.0 / 3
	if summary.SuccessRate < expectedRate-0.01 || summary.SuccessRate > expectedRate+0.01 {
		t.Errorf("unexpected success rate: %f", summary.SuccessRate)
	}
	if summary.Jobs[0].LastStatus != "success" {
		t.Errorf("expected last status success, got %s", summary.Jobs[0].LastStatus)
	}
}

func TestMetrics_MethodNotAllowed(t *testing.T) {
	store := tempMetricsStore(t)
	h := NewMetricsHandler(store, []string{})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
