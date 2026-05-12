package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/history"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "history.json")
}

func TestRecord_And_Last(t *testing.T) {
	s, err := history.NewStore(tempPath(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	e := history.Entry{JobName: "backup", Success: true, Timestamp: time.Now()}
	if err := s.Record(e); err != nil {
		t.Fatalf("Record: %v", err)
	}

	got, ok := s.Last("backup")
	if !ok {
		t.Fatal("expected entry, got none")
	}
	if got.JobName != "backup" || !got.Success {
		t.Errorf("unexpected entry: %+v", got)
	}
}

func TestLast_UnknownJob(t *testing.T) {
	s, _ := history.NewStore(tempPath(t))
	_, ok := s.Last("nonexistent")
	if ok {
		t.Error("expected no entry for unknown job")
	}
}

func TestAll_MultipleEntries(t *testing.T) {
	s, _ := history.NewStore(tempPath(t))
	for i := 0; i < 3; i++ {
		_ = s.Record(history.Entry{JobName: "job", Success: i%2 == 0, Timestamp: time.Now()})
	}
	if len(s.All()) != 3 {
		t.Errorf("expected 3 entries, got %d", len(s.All()))
	}
}

func TestPersistence_ReloadFromDisk(t *testing.T) {
	path := tempPath(t)
	s1, _ := history.NewStore(path)
	_ = s1.Record(history.Entry{JobName: "sync", Success: false, Message: "exit 1", Timestamp: time.Now()})

	s2, err := history.NewStore(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got, ok := s2.Last("sync")
	if !ok {
		t.Fatal("entry not persisted")
	}
	if got.Message != "exit 1" {
		t.Errorf("message mismatch: %q", got.Message)
	}
}

func TestNewStore_MissingFile_OK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does_not_exist.json")
	_, err := history.NewStore(path)
	if err != nil {
		t.Errorf("expected no error for missing file, got: %v", err)
	}
}

func TestNewStore_CorruptFile_Error(t *testing.T) {
	path := tempPath(t)
	os.WriteFile(path, []byte("not json{"), 0o644)
	_, err := history.NewStore(path)
	if err == nil {
		t.Error("expected error for corrupt file")
	}
}
