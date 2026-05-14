package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/cronwatch/internal/history"
)

// MetricsSummary holds aggregated statistics for all monitored jobs.
type MetricsSummary struct {
	GeneratedAt   time.Time         `json:"generated_at"`
	TotalJobs     int               `json:"total_jobs"`
	TotalRuns     int               `json:"total_runs"`
	TotalFailures int               `json:"total_failures"`
	SuccessRate   float64           `json:"success_rate_pct"`
	Jobs          []JobMetrics      `json:"jobs"`
}

// JobMetrics holds per-job statistics.
type JobMetrics struct {
	Name          string  `json:"name"`
	TotalRuns     int     `json:"total_runs"`
	Failures      int     `json:"failures"`
	SuccessRate   float64 `json:"success_rate_pct"`
	LastStatus    string  `json:"last_status"`
	LastRunAt     string  `json:"last_run_at,omitempty"`
}

// MetricsHandler serves aggregated job metrics.
type MetricsHandler struct {
	store history.Storer
	jobs  []string
}

// NewMetricsHandler creates a MetricsHandler for the given job names.
func NewMetricsHandler(store history.Storer, jobs []string) *MetricsHandler {
	return &MetricsHandler{store: store, jobs: jobs}
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	summary := MetricsSummary{
		GeneratedAt: time.Now().UTC(),
		TotalJobs:   len(h.jobs),
	}

	for _, name := range h.jobs {
		entries, err := h.store.All(name)
		if err != nil || len(entries) == 0 {
			summary.Jobs = append(summary.Jobs, JobMetrics{Name: name, LastStatus: "unknown"})
			continue
		}

		var failures int
		for _, e := range entries {
			if !e.Success {
				failures++
			}
		}

		total := len(entries)
		last := entries[len(entries)-1]
		status := "success"
		if !last.Success {
			status = "failure"
		}

		rate := 0.0
		if total > 0 {
			rate = float64(total-failures) / float64(total) * 100
		}

		summary.TotalRuns += total
		summary.TotalFailures += failures
		summary.Jobs = append(summary.Jobs, JobMetrics{
			Name:        name,
			TotalRuns:   total,
			Failures:    failures,
			SuccessRate: rate,
			LastStatus:  status,
			LastRunAt:   last.RunAt.UTC().Format(time.RFC3339),
		})
	}

	if summary.TotalRuns > 0 {
		summary.SuccessRate = float64(summary.TotalRuns-summary.TotalFailures) / float64(summary.TotalRuns) * 100
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
