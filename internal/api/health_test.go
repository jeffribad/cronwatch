package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealth_OK(t *testing.T) {
	h := NewHealthHandler("v1.2.3")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}
	if resp["version"] != "v1.2.3" {
		t.Errorf("expected version v1.2.3, got %q", resp["version"])
	}
	if !strings.HasPrefix(resp["go_version"], "go") {
		t.Errorf("unexpected go_version: %q", resp["go_version"])
	}
	if resp["uptime"] == "" {
		t.Error("expected non-empty uptime")
	}
}

func TestHealth_UptimeIncreases(t *testing.T) {
	h := NewHealthHandler("dev")
	h.startTime = time.Now().Add(-5 * time.Second)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var resp map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	// uptime should be at least "5s"
	if resp["uptime"] == "0s" || resp["uptime"] == "" {
		t.Errorf("expected uptime > 0, got %q", resp["uptime"])
	}
}

func TestHealth_MethodNotAllowed(t *testing.T) {
	h := NewHealthHandler("v0.0.1")

	for _, method := range []string{http.MethodPost, http.MethodDelete, http.MethodPut} {
		req := httptest.NewRequest(method, "/healthz", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, rec.Code)
		}
	}
}
