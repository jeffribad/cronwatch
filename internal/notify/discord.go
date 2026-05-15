package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// DiscordNotifier sends alert messages to a Discord channel via webhook.
type DiscordNotifier struct {
	webhookURL string
	client     *http.Client
}

type discordPayload struct {
	Content string `json:"content"`
}

// NewDiscordNotifier creates a new DiscordNotifier with the given webhook URL.
func NewDiscordNotifier(webhookURL string) *DiscordNotifier {
	return &DiscordNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

// Send posts the alert message to the configured Discord webhook.
func (d *DiscordNotifier) Send(subject, body string) error {
	payload := discordPayload{
		Content: fmt.Sprintf("**%s**\n%s", subject, body),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, d.webhookURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("discord: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("discord: send request: %w", err)
	}
	defer resp.Body.Close()

	// Discord returns 204 No Content on success
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord: unexpected status %d", resp.StatusCode)
	}

	return nil
}
