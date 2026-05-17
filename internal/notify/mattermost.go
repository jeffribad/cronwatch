package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// MattermostNotifier sends alerts to a Mattermost incoming webhook.
type MattermostNotifier struct {
	webhookURL string
	channel    string
	username   string
	client     *http.Client
}

type mattermostPayload struct {
	Text     string `json:"text"`
	Channel  string `json:"channel,omitempty"`
	Username string `json:"username,omitempty"`
}

// NewMattermostNotifier creates a new MattermostNotifier.
// channel and username are optional; leave empty to use webhook defaults.
func NewMattermostNotifier(webhookURL, channel, username string) *MattermostNotifier {
	return &MattermostNotifier{
		webhookURL: webhookURL,
		channel:    channel,
		username:   username,
		client:     &http.Client{},
	}
}

// Send posts the alert message to Mattermost.
func (m *MattermostNotifier) Send(subject, body string) error {
	payload := mattermostPayload{
		Text:     fmt.Sprintf("**%s**\n%s", subject, body),
		Channel:  m.channel,
		Username: m.username,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mattermost: marshal payload: %w", err)
	}

	resp, err := m.client.Post(m.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("mattermost: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mattermost: unexpected status %d", resp.StatusCode)
	}
	return nil
}
