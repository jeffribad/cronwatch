package scheduler_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/alert"
	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/history"
	"github.com/user/cronwatch/internal/monitor"
	"github.com/user/cronwatch/internal/scheduler"
)

func makeTestConfig(intervalSec int, jobs []config.Job) *config.Config {
	return &config.Config{
		CheckIntervalSeconds: intervalSec,
		Jobs:                 jobs,
	}
}

func TestScheduler_StartsAndStops(t *testing.T) {
	cfg := makeTestConfig(1, []config.Job{
		{Name: "test-job", MaxAgeSeconds: 300},
	})

	dir := t.TempDir()
	store, err := history.NewStore(dir + "/history.json")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	logger := log.New(os.Stderr, "", 0)
	dispatcher := alert.NewDispatcher(nil, logger)
	mon := monitor.New(cfg, store, dispatcher, logger)

	sched := scheduler.New(cfg, mon, logger)
	sched.Start()

	time.Sleep(50 * time.Millisecond)
	sched.Stop() // should not hang
}

func TestScheduler_RunsChecks(t *testing.T) {
	cfg := makeTestConfig(1, []config.Job{
		{Name: "job-a", MaxAgeSeconds: 1},
	})

	dir := t.TempDir()
	store, err := history.NewStore(dir + "/history.json")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	logger := log.New(os.Stderr, "", 0)
	dispatcher := alert.NewDispatcher(nil, logger)
	mon := monitor.New(cfg, store, dispatcher, logger)

	// Record a recent run so no alert fires.
	if err := mon.RecordRun("job-a", true, ""); err != nil {
		t.Fatalf("RecordRun: %v", err)
	}

	sched := scheduler.New(cfg, mon, logger)
	sched.Start()
	time.Sleep(150 * time.Millisecond)
	sched.Stop()
}
