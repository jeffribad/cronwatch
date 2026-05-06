package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "cronwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
check_interval: 30s
alerts:
  email: ops@example.com
jobs:
  - name: backup
    schedule: "0 2 * * *"
    timeout: 10m
    command: /usr/local/bin/backup.sh
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 30*time.Second {
		t.Errorf("expected 30s interval, got %v", cfg.CheckInterval)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("expected job name 'backup', got %q", cfg.Jobs[0].Name)
	}
}

func TestLoad_DefaultInterval(t *testing.T) {
	path := writeTempConfig(t, `
jobs:
  - name: cleanup
    schedule: "*/5 * * * *"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 60*time.Second {
		t.Errorf("expected default 60s, got %v", cfg.CheckInterval)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_NoJobs(t *testing.T) {
	path := writeTempConfig(t, `check_interval: 10s\n`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty jobs")
	}
}

func TestLoad_JobMissingSchedule(t *testing.T) {
	path := writeTempConfig(t, `
jobs:
  - name: broken
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing schedule")
	}
}
