package retention_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/history"
	"github.com/user/cronwatch/internal/retention"
)

func tempStore(t *testing.T) *history.Store {
	t.Helper()
	f, err := os.CreateTemp("", "retention-test-*.json")
	if err != nil {
		t.Fatalf("tempStore: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	store, err := history.NewStore(f.Name())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return store
}

func TestPurge_RemovesOldEntries(t *testing.T) {
	store := tempStore(t)

	oldTime := time.Now().Add(-48 * time.Hour)
	newTime := time.Now().Add(-1 * time.Hour)

	_ = store.Record("job1", history.Entry{JobName: "job1", Success: true, Timestamp: oldTime})
	_ = store.Record("job1", history.Entry{JobName: "job1", Success: true, Timestamp: newTime})

	policy := retention.New(store, 24*time.Hour, log.Default())
	removed, err := policy.Purge()
	if err != nil {
		t.Fatalf("Purge: %v", err)
	}
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	entries, _ := store.All()
	if len(entries["job1"]) != 1 {
		t.Errorf("expected 1 remaining entry, got %d", len(entries["job1"]))
	}
}

func TestPurge_KeepsRecentEntries(t *testing.T) {
	store := tempStore(t)

	_ = store.Record("job2", history.Entry{JobName: "job2", Success: true, Timestamp: time.Now()})

	policy := retention.New(store, 24*time.Hour, log.Default())
	removed, err := policy.Purge()
	if err != nil {
		t.Fatalf("Purge: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestPurge_EmptyStore(t *testing.T) {
	store := tempStore(t)
	policy := retention.New(store, 24*time.Hour, nil)
	removed, err := policy.Purge()
	if err != nil {
		t.Fatalf("Purge on empty store: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestRunEvery_PurgesOnTick(t *testing.T) {
	store := tempStore(t)
	old := time.Now().Add(-72 * time.Hour)
	_ = store.Record("job3", history.Entry{JobName: "job3", Success: false, Timestamp: old})

	policy := retention.New(store, 24*time.Hour, log.Default())
	quit := make(chan struct{})
	policy.RunEvery(50*time.Millisecond, quit)

	time.Sleep(120 * time.Millisecond)
	close(quit)

	entries, _ := store.All()
	if len(entries["job3"]) != 0 {
		t.Errorf("expected entries purged, got %d", len(entries["job3"]))
	}
}
