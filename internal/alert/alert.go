package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Notifier sends alert notifications.
type Notifier interface {
	Send(subject, body string) error
}

// WebhookNotifier posts alerts to an HTTP endpoint.
type WebhookNotifier struct {
	URL     string
	Timeout time.Duration
}

type webhookPayload struct {
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Timestamp string `json:"timestamp"`
}

// NewWebhookNotifier creates a WebhookNotifier with the given URL.
func NewWebhookNotifier(url string, timeout time.Duration) *WebhookNotifier {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookNotifier{URL: url, Timeout: timeout}
}

// Send posts a JSON payload to the configured webhook URL.
func (w *WebhookNotifier) Send(subject, body string) error {
	payload := webhookPayload{
		Subject:   subject,
		Body:      body,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("alert: marshal payload: %w", err)
	}

	client := &http.Client{Timeout: w.Timeout}
	resp, err := client.Post(w.URL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("alert: post webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alert: webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// LogNotifier writes alerts to a provided writer (e.g. os.Stderr).
type LogNotifier struct {
	Writer interface{ WriteString(string) (int, error) }
}

// Send writes the alert as a formatted string.
func (l *LogNotifier) Send(subject, body string) error {
	msg := fmt.Sprintf("[ALERT] %s | %s | %s\n",
		time.Now().UTC().Format(time.RFC3339), subject, body)
	_, err := l.Writer.WriteString(msg)
	return err
}
