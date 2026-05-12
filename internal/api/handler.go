package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/user/cronwatch/internal/history"
	"github.com/user/cronwatch/internal/monitor"
)

// Handler exposes HTTP endpoints for cronwatch.
type Handler struct {
	mon    *monitor.Monitor
	store  *history.Store
	logger *log.Logger
}

// NewHandler creates an HTTP handler wired to the given monitor and store.
func NewHandler(mon *monitor.Monitor, store *history.Store, logger *log.Logger) *Handler {
	return &Handler{mon: mon, store: store, logger: logger}
}

// RegisterRoutes attaches all routes to mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/api/report", h.report)
	mux.HandleFunc("/api/jobs", h.listJobs)
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type runReport struct {
	Job       string    `json:"job"`
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job     string `json:"job"`
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Job == "" {
		http.Error(w, "job name required", http.StatusBadRequest)
		return
	}
	if err := h.mon.RecordRun(req.Job, req.Success, req.Message); err != nil {
		h.logger.Printf("api: RecordRun error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	all, err := h.store.All()
	if err != nil {
		h.logger.Printf("api: All error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(all)
}
