package api

import (
	"net/http"

	"github.com/yourorg/cronwatch/internal/config"
	"github.com/yourorg/cronwatch/internal/history"
	"github.com/yourorg/cronwatch/internal/monitor"
)

// RegisterRoutes wires all HTTP handlers onto mux.
// version is the build version string surfaced by the health endpoint.
func RegisterRoutes(
	mux *http.ServeMux,
	cfg *config.Config,
	store *history.Store,
	mon *monitor.Monitor,
	version string,
) {
	// Legacy simple health check (plain 200 OK).
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Detailed health / build-info endpoint.
	mux.Handle("/health", NewHealthHandler(version))

	// Job report ingestion.
	mux.Handle("/report", NewHandler(store, mon))

	// Per-job history.
	mux.Handle("/history", newHistoryHandler(store))

	// Aggregated status per job.
	mux.Handle("/status", NewStatusHandler(store, cfg))

	// Prometheus-style metrics.
	mux.Handle("/metrics", NewMetricsHandler(store))
}
