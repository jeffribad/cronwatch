package monitor

import (
	"log"
	"sync"
	"time"

	"cronwatch/internal/config"
)

// JobState tracks the last known state of a monitored cron job.
type JobState struct {
	Name        string
	LastSeen    time.Time
	Missed      bool
	FailCount   int
}

// Monitor watches cron job execution by scanning logs and tracking state.
type Monitor struct {
	cfg    *config.Config
	states map[string]*JobState
	mu     sync.Mutex
	alert  AlertFunc
}

// AlertFunc is called when a job is missed or fails.
type AlertFunc func(job string, reason string)

// New creates a new Monitor with the given config and alert callback.
func New(cfg *config.Config, alertFn AlertFunc) *Monitor {
	states := make(map[string]*JobState, len(cfg.Jobs))
	for _, j := range cfg.Jobs {
		states[j.Name] = &JobState{Name: j.Name}
	}
	return &Monitor{
		cfg:    cfg,
		states: states,
		alert:  alertFn,
	}
}

// RecordRun updates the last-seen time for a job and resets missed/fail state.
func (m *Monitor) RecordRun(name string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.states[name]
	if !ok {
		log.Printf("[monitor] unknown job: %s", name)
		return
	}
	state.LastSeen = time.Now()
	state.Missed = false
	if !success {
		state.FailCount++
		m.alert(name, "job exited with failure")
	} else {
		state.FailCount = 0
	}
}

// Check evaluates all jobs and fires alerts for any that have exceeded their interval.
func (m *Monitor) Check() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for _, j := range m.cfg.Jobs {
		state := m.states[j.Name]
		if state.LastSeen.IsZero() {
			continue
		}
		deadline := state.LastSeen.Add(time.Duration(j.IntervalMinutes) * time.Minute)
		if now.After(deadline) && !state.Missed {
			state.Missed = true
			m.alert(j.Name, "job missed expected run window")
		}
	}
}

// State returns a copy of the current state for a job.
func (m *Monitor) State(name string) (JobState, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.states[name]
	if !ok {
		return JobState{}, false
	}
	return *s, true
}
