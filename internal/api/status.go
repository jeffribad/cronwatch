package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/cronwatch/internal/history"
	"github.com/user/cronwatch/internal/config"
)

// JobStatus represents the current status summary of a single cron job.
type JobStatus struct {
	Name        string     `json:"name"`
	Schedule    string     `json:"schedule"`
	LastRun     *time.Time `json:"last_run,omitempty"`
	LastSuccess *time.Time `json:"last_success,omitempty"`
	LastFailure *time.Time `json:"last_failure,omitempty"`
	Status      string     `json:"status"`
}

// StatusHandler returns a summary of all configured jobs and their last known state.
type StatusHandler struct {
	store  *history.Store
	cfg    *config.Config
}

// NewStatusHandler creates a new StatusHandler.
func NewStatusHandler(store *history.Store, cfg *config.Config) *StatusHandler {
	return &StatusHandler{store: store, cfg: cfg}
}

func (h *StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statuses := make([]JobStatus, 0, len(h.cfg.Jobs))

	for _, job := range h.cfg.Jobs {
		js := JobStatus{
			Name:     job.Name,
			Schedule: job.Schedule,
			Status:   "unknown",
		}

		if last, err := h.store.Last(job.Name); err == nil {
			js.LastRun = &last.Time
			if last.Success {
				js.LastSuccess = &last.Time
				js.Status = "ok"
			} else {
				js.LastFailure = &last.Time
				js.Status = "failing"
			}
		}

		statuses = append(statuses, js)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}
