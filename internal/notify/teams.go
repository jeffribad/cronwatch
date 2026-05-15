package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// TeamsNotifier sends alerts to a Microsoft Teams channel via an incoming webhook.
type TeamsNotifier struct {
	webhookURL string
	client     *http.Client
}

// teamsPayload is the message card format accepted by Teams webhooks.
type teamsPayload struct {
	Type       string `json:"@type"`
	Context    string `json:"@context"`
	ThemeColor string `json:"themeColor"`
	Summary    string `json:"summary"`
	Text       string `json:"text"`
}

// NewTeamsNotifier creates a TeamsNotifier that posts to the given webhook URL.
func NewTeamsNotifier(webhookURL string) *TeamsNotifier {
	return &TeamsNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

// Send delivers the alert message to the configured Teams channel.
func (t *TeamsNotifier) Send(subject, body string) error {
	payload := teamsPayload{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: "FF0000",
		Summary:    subject,
		Text:       fmt.Sprintf("**%s**\n\n%s", subject, body),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("teams: marshal payload: %w", err)
	}

	resp, err := t.client.Post(t.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("teams: post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("teams: unexpected status %d", resp.StatusCode)
	}

	return nil
}
