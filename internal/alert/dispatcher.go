package alert

import (
	"fmt"
	"log"
)

// Dispatcher fans out an alert to one or more Notifiers.
type Dispatcher struct {
	notifiers []Notifier
	logger    *log.Logger
}

// NewDispatcher creates a Dispatcher with the given notifiers.
func NewDispatcher(logger *log.Logger, notifiers ...Notifier) *Dispatcher {
	return &Dispatcher{
		notifiers: notifiers,
		logger:    logger,
	}
}

// Dispatch sends the alert to all registered notifiers.
// It continues even if one notifier fails, collecting all errors.
func (d *Dispatcher) Dispatch(subject, body string) error {
	var errs []error
	for _, n := range d.notifiers {
		if err := n.Send(subject, body); err != nil {
			d.logger.Printf("alert dispatch error: %v", err)
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("dispatch: %d notifier(s) failed (first: %w)", len(errs), errs[0])
	}
	return nil
}

// AlertJobFailure sends a standardised failure alert for a cron job.
func (d *Dispatcher) AlertJobFailure(jobName string, exitCode int) error {
	subject := fmt.Sprintf("cronwatch: job '%s' failed", jobName)
	body := fmt.Sprintf("Job '%s' exited with code %d.", jobName, exitCode)
	return d.Dispatch(subject, body)
}

// AlertMissedRun sends a standardised missed-run alert for a cron job.
func (d *Dispatcher) AlertMissedRun(jobName, expectedInterval string) error {
	subject := fmt.Sprintf("cronwatch: job '%s' missed scheduled run", jobName)
	body := fmt.Sprintf("Job '%s' has not run within the expected interval of %s.", jobName, expectedInterval)
	return d.Dispatch(subject, body)
}
