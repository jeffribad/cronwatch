package alert_test

import (
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/yourorg/cronwatch/internal/alert"
)

type stubNotifier struct {
	calls []string
	err   error
}

func (s *stubNotifier) Send(subject, body string) error {
	s.calls = append(s.calls, subject+"|"+body)
	return s.err
}

func newLogger() *log.Logger {
	return log.New(os.Stderr, "", 0)
}

func TestDispatcher_Dispatch_AllSucceed(t *testing.T) {
	a, b := &stubNotifier{}, &stubNotifier{}
	d := alert.NewDispatcher(newLogger(), a, b)

	if err := d.Dispatch("subj", "body"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(a.calls) != 1 || len(b.calls) != 1 {
		t.Errorf("expected each notifier called once")
	}
}

func TestDispatcher_Dispatch_OneFailsContinues(t *testing.T) {
	bad := &stubNotifier{err: errors.New("network down")}
	good := &stubNotifier{}
	d := alert.NewDispatcher(newLogger(), bad, good)

	err := d.Dispatch("subj", "body")
	if err == nil {
		t.Fatal("expected error when a notifier fails")
	}
	// good notifier should still have been called
	if len(good.calls) != 1 {
		t.Error("expected good notifier to be called despite earlier failure")
	}
}

func TestDispatcher_Dispatch_NoNotifiers(t *testing.T) {
	d := alert.NewDispatcher(newLogger())

	if err := d.Dispatch("subj", "body"); err != nil {
		t.Fatalf("expected no error with zero notifiers, got %v", err)
	}
}

func TestDispatcher_AlertJobFailure(t *testing.T) {
	n := &stubNotifier{}
	d := alert.NewDispatcher(newLogger(), n)

	if err := d.AlertJobFailure("backup-db", 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(n.calls) != 1 {
		t.Fatal("expected one call")
	}
	if !strings.Contains(n.calls[0], "backup-db") {
		t.Errorf("expected job name in alert: %q", n.calls[0])
	}
}

func TestDispatcher_AlertMissedRun(t *testing.T) {
	n := &stubNotifier{}
	d := alert.NewDispatcher(newLogger(), n)

	if err := d.AlertMissedRun("nightly-report", "24h"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(n.calls[0], "nightly-report") {
		t.Errorf("expected job name in alert: %q", n.calls[0])
	}
	if !strings.Contains(n.calls[0], "24h") {
		t.Errorf("expected interval in alert: %q", n.calls[0])
	}
}

func TestDispatcher_Dispatch_AllFail(t *testing.T) {
	a := &stubNotifier{err: errors.New("timeout")}
	b := &stubNotifier{err: errors.New("refused")}
	d := alert.NewDispatcher(newLogger(), a, b)

	err := d.Dispatch("subj", "body")
	if err == nil {
		t.Fatal("expected error when all notifiers fail")
	}
	// both notifiers should have been attempted
	if len(a.calls) != 1 || len(b.calls) != 1 {
		t.Errorf("expected both notifiers to be attempted, got a=%d b=%d", len(a.calls), len(b.calls))
	}
}
