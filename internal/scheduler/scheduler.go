package scheduler

import (
	"log"
	"sync"
	"time"

	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/monitor"
)

// Scheduler periodically checks all configured cron jobs for missed runs.
type Scheduler struct {
	cfg     *config.Config
	mon     *monitor.Monitor
	logger  *log.Logger
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// New creates a new Scheduler.
func New(cfg *config.Config, mon *monitor.Monitor, logger *log.Logger) *Scheduler {
	return &Scheduler{
		cfg:    cfg,
		mon:    mon,
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

// Start begins the scheduling loop in a background goroutine.
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		interval := time.Duration(s.cfg.CheckIntervalSeconds) * time.Second
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		s.logger.Printf("scheduler: starting with interval %s", interval)
		for {
			select {
			case <-ticker.C:
				s.runChecks()
			case <-s.stopCh:
				s.logger.Println("scheduler: stopping")
				return
			}
		}
	}()
}

// Stop signals the scheduling loop to exit and waits for it to finish.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

func (s *Scheduler) runChecks() {
	for _, job := range s.cfg.Jobs {
		if err := s.mon.Check(job.Name); err != nil {
			s.logger.Printf("scheduler: check error for job %q: %v", job.Name, err)
		}
	}
}
