package retention

import (
	"log"
	"time"

	"github.com/user/cronwatch/internal/history"
)

// Policy defines how long job history entries are retained.
type Policy struct {
	MaxAge time.Duration
	store  *history.Store
	logger *log.Logger
}

// New creates a new retention Policy with the given max age.
func New(store *history.Store, maxAge time.Duration, logger *log.Logger) *Policy {
	if logger == nil {
		logger = log.Default()
	}
	return &Policy{
		MaxAge: maxAge,
		store:  store,
		logger: logger,
	}
}

// Purge removes history entries older than the policy's MaxAge.
// Returns the number of entries removed.
func (p *Policy) Purge() (int, error) {
	entries, err := p.store.All()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-p.MaxAge)
	removed := 0

	for jobName, records := range entries {
		var kept []history.Entry
		for _, rec := range records {
			if rec.Timestamp.After(cutoff) {
				kept = append(kept, rec)
			} else {
				removed++
			}
		}
		if err := p.store.Replace(jobName, kept); err != nil {
			p.logger.Printf("retention: failed to replace entries for %s: %v", jobName, err)
		}
	}

	if removed > 0 {
		p.logger.Printf("retention: purged %d expired entries (older than %s)", removed, p.MaxAge)
	}
	return removed, nil
}

// RunEvery starts a background goroutine that runs Purge on the given interval.
// Stop the loop by closing the quit channel.
func (p *Policy) RunEvery(interval time.Duration, quit <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if _, err := p.Purge(); err != nil {
					p.logger.Printf("retention: purge error: %v", err)
				}
			case <-quit:
				return
			}
		}
	}()
}
