package history

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Entry represents a single recorded execution of a cron job.
type Entry struct {
	JobName   string        `json:"job_name"`
	Success   bool          `json:"success"`
	Message   string        `json:"message,omitempty"`
	Duration  time.Duration `json:"duration_ns,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// Store persists job execution history to a JSON file.
type Store struct {
	mu      sync.RWMutex
	path    string
	records map[string][]Entry
}

// NewStore loads or initialises a history store at the given path.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path, records: make(map[string][]Entry)}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("history: read %s: %w", path, err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &s.records); err != nil {
			return nil, fmt.Errorf("history: parse %s: %w", path, err)
		}
	}
	return s, nil
}

// Record appends an entry for the named job and flushes to disk.
func (s *Store) Record(jobName string, e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[jobName] = append(s.records[jobName], e)
	return s.flush()
}

// Last returns the most recent entry for the given job, or false if none.
func (s *Store) Last(jobName string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries := s.records[jobName]
	if len(entries) == 0 {
		return Entry{}, false
	}
	return entries[len(entries)-1], true
}

// All returns a shallow copy of all recorded entries keyed by job name.
func (s *Store) All() (map[string][]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]Entry, len(s.records))
	for k, v := range s.records {
		cp := make([]Entry, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out, nil
}

// Replace overwrites the stored entries for a single job and flushes to disk.
func (s *Store) Replace(jobName string, entries []Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(entries) == 0 {
		delete(s.records, jobName)
	} else {
		s.records[jobName] = entries
	}
	return s.flush()
}

// flush writes the current records to disk. Caller must hold s.mu.
func (s *Store) flush() error {
	data, err := json.MarshalIndent(s.records, "", "  ")
	if err != nil {
		return fmt.Errorf("history: marshal: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("history: write %s: %w", s.path, err)
	}
	return nil
}
