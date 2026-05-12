package history

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry represents a single recorded execution of a cron job.
type Entry struct {
	JobName   string    `json:"job_name"`
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// Store persists job run history to a JSON file.
type Store struct {
	mu      sync.RWMutex
	path    string
	entries []Entry
}

// NewStore creates a Store backed by the given file path.
// Existing entries are loaded from disk if the file exists.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Record appends a new entry and flushes to disk.
func (s *Store) Record(e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, e)
	return s.save()
}

// Last returns the most recent entry for the given job name, or false if none.
func (s *Store) Last(jobName string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.entries) - 1; i >= 0; i-- {
		if s.entries[i].JobName == jobName {
			return s.entries[i], true
		}
	}
	return Entry{}, false
}

// All returns a copy of all stored entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, len(s.entries))
	copy(out, s.entries)
	return out
}

func (s *Store) load() error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&s.entries)
}

func (s *Store) save() error {
	f, err := os.CreateTemp("", "cronwatch-history-*")
	if err != nil {
		return err
	}
	tmpName := f.Name()
	if err := json.NewEncoder(f).Encode(s.entries); err != nil {
		f.Close()
		os.Remove(tmpName)
		return err
	}
	f.Close()
	return os.Rename(tmpName, s.path)
}
