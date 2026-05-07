package monitor

import (
	"testing"
	"time"

	"cronwatch/internal/config"
)

func makeConfig(intervalMinutes int) *config.Config {
	return &config.Config{
		Jobs: []config.Job{
			{Name: "backup", IntervalMinutes: intervalMinutes},
		},
	}
}

func TestRecordRun_Success(t *testing.T) {
	var alerted bool
	m := New(makeConfig(60), func(job, reason string) { alerted = true })
	m.RecordRun("backup", true)

	s, ok := m.State("backup")
	if !ok {
		t.Fatal("expected state for backup")
	}
	if s.LastSeen.IsZero() {
		t.Error("expected LastSeen to be set")
	}
	if s.FailCount != 0 {
		t.Errorf("expected FailCount 0, got %d", s.FailCount)
	}
	if alerted {
		t.Error("expected no alert on success")
	}
}

func TestRecordRun_Failure(t *testing.T) {
	var alerts []string
	m := New(makeConfig(60), func(job, reason string) { alerts = append(alerts, reason) })
	m.RecordRun("backup", false)

	s, _ := m.State("backup")
	if s.FailCount != 1 {
		t.Errorf("expected FailCount 1, got %d", s.FailCount)
	}
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(alerts))
	}
}

func TestCheck_MissedJob(t *testing.T) {
	var missedJob string
	m := New(makeConfig(1), func(job, reason string) { missedJob = job })

	// Seed a last-seen time far in the past
	m.mu.Lock()
	m.states["backup"].LastSeen = time.Now().Add(-2 * time.Minute)
	m.mu.Unlock()

	m.Check()

	if missedJob != "backup" {
		t.Errorf("expected missed alert for backup, got %q", missedJob)
	}
	s, _ := m.State("backup")
	if !s.Missed {
		t.Error("expected Missed to be true")
	}
}

func TestCheck_NoAlertWhenRecent(t *testing.T) {
	var alerted bool
	m := New(makeConfig(60), func(job, reason string) { alerted = true })
	m.RecordRun("backup", true)
	m.Check()

	if alerted {
		t.Error("expected no alert for recently seen job")
	}
}

func TestCheck_NoAlertWhenNeverSeen(t *testing.T) {
	var alerted bool
	m := New(makeConfig(1), func(job, reason string) { alerted = true })
	m.Check()
	if alerted {
		t.Error("expected no alert for job never recorded")
	}
}
