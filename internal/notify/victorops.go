package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultVictorOpsBaseURL = "https://alert.victorops.com/integrations/generic/20131114/alert"

// VictorOpsNotifier sends alerts to VictorOps (Splunk On-Call) via the REST endpoint.
type VictorOpsNotifier struct {
	apiKey   string
	routingKey string
	baseURL  string
	client   *http.Client
}

type victorOpsPayload struct {
	MessageType       string `json:"message_type"`
	EntityID          string `json:"entity_id"`
	EntityDisplayName string `json:"entity_display_name"`
	StateMessage      string `json:"state_message"`
	Timestamp         int64  `json:"timestamp"`
}

// NewVictorOpsNotifier creates a VictorOpsNotifier with the given API and routing keys.
func NewVictorOpsNotifier(apiKey, routingKey string) *VictorOpsNotifier {
	return &VictorOpsNotifier{
		apiKey:     apiKey,
		routingKey: routingKey,
		baseURL:    defaultVictorOpsBaseURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Send delivers an alert message to VictorOps.
func (v *VictorOpsNotifier) Send(subject, body string) error {
	payload := victorOpsPayload{
		MessageType:       "CRITICAL",
		EntityID:          subject,
		EntityDisplayName: subject,
		StateMessage:      body,
		Timestamp:         time.Now().Unix(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("victorops: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", v.baseURL, v.apiKey, v.routingKey)
	resp, err := v.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("victorops: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("victorops: unexpected status %d", resp.StatusCode)
	}
	return nil
}
