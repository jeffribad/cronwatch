package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OpsGenieNotifier sends alerts to OpsGenie via the Alerts API.
type OpsGenieNotifier struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

type opsGeniePayload struct {
	Message     string            `json:"message"`
	Description string            `json:"description"`
	Priority    string            `json:"priority"`
	Tags        []string          `json:"tags,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
}

// NewOpsGenieNotifier creates a new OpsGenieNotifier.
// apiKey is the OpsGenie API key; baseURL can be overridden for testing.
func NewOpsGenieNotifier(apiKey, baseURL string) *OpsGenieNotifier {
	if baseURL == "" {
		baseURL = "https://api.opsgenie.com/v2/alerts"
	}
	return &OpsGenieNotifier{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Send delivers an alert message to OpsGenie.
func (o *OpsGenieNotifier) Send(subject, body string) error {
	payload := opsGeniePayload{
		Message:     subject,
		Description: body,
		Priority:    "P2",
		Tags:        []string{"cronwatch"},
		Details:     map[string]string{"source": "cronwatch"},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("opsgenie: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, o.baseURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("opsgenie: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("opsgenie: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opsgenie: unexpected status %d", resp.StatusCode)
	}
	return nil
}
