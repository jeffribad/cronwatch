package api

import (
	"net/http"

	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/history"
	"github.com/user/cronwatch/internal/monitor"
)

// RegisterRoutes wires all API handlers onto the given mux.
func RegisterRoutes(
	mux *http.ServeMux,
	cfg *config.Config,
	store history.Storer,
	mon *monitor.Monitor,
) {
	jobNames := make([]string, 0, len(cfg.Jobs))
	for _, j := range cfg.Jobs {
		jobNames = append(jobNames, j.Name)
	}

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Job report endpoint (POST /report?job=name&success=true)
	reportHandler := NewHandler(store, mon)
	mux.Handle("/report", reportHandler)

	// Per-job run history
	historyHandler := NewHistoryHandler(store)
	mux.Handle("/history", historyHandler)

	// Current status of all jobs
	statusHandler := NewStatusHandler(store, cfg)
	mux.Handle("/status", statusHandler)

	// Aggregated metrics
	metricsHandler := NewMetricsHandler(store, jobNames)
	mux.Handle("/metrics", metricsHandler)
}
