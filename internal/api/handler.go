package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/cronwatch/internal/history"
	"github.com/user/cronwatch/internal/monitor"
)

// Handler holds dependencies for HTTP endpoints.
type Handler struct {
	monitor *monitor.Monitor
	store   *history.Store
}

// NewHandler constructs a Handler.
func NewHandler(m *monitor.Monitor, s *history.Store) *Handler {
	return &Handler{monitor: m, store: s}
}

// RegisterRoutes attaches all routes to the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.Healthz)
	mux.HandleFunc("/report", h.Report)
	mux.HandleFunc("/history", h.History)
}

// Healthz returns a simple 200 OK.
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type reportRequest struct {
	Job     string `json:"job"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Report accepts an inbound job execution result and records it.
func (h *Handler) Report(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req reportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	h.monitor.RecordRun(req.Job, req.Success, req.Message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"recorded": req.Job})
}

type historyEntry struct {
	JobName   string    `json:"job_name"`
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// toHistoryEntry converts a history.Entry to the API historyEntry type.
func toHistoryEntry(e history.Entry) historyEntry {
	return historyEntry{
		JobName:   e.JobName,
		Success:   e.Success,
		Message:   e.Message,
		Timestamp: e.Timestamp,
	}
}

// History returns all stored execution history.
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	all, err := h.store.All()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	out := make(map[string][]historyEntry, len(all))
	for job, entries := range all {
		for _, e := range entries {
			out[job] = append(out[job], toHistoryEntry(e))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
