package api

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// HealthHandler serves detailed health and build info.
type HealthHandler struct {
	startTime time.Time
	version  string
}

// NewHealthHandler creates a HealthHandler. version is the build version string.
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:  version,
	}
}

type healthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Uptime    string `json:"uptime"`
	StartedAt string `json:"started_at"`
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := healthResponse{
		Status:    "ok",
		Version:   h.version,
		GoVersion: runtime.Version(),
		Uptime:    time.Since(h.startTime).Round(time.Second).String(),
		StartedAt: h.startTime.UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
