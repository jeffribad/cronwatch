package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// RocketChatNotifier sends alert messages to a Rocket.Chat incoming webhook.
type RocketChatNotifier struct {
	webhookURL string
	channel    string
	username   string
	client     *http.Client
}

type rocketChatPayload struct {
	Text     string `json:"text"`
	Channel  string `json:"channel,omitempty"`
	Username string `json:"username,omitempty"`
}

// NewRocketChatNotifier creates a new RocketChatNotifier.
// channel and username are optional; pass empty strings to use webhook defaults.
func NewRocketChatNotifier(webhookURL, channel, username string) *RocketChatNotifier {
	return &RocketChatNotifier{
		webhookURL: webhookURL,
		channel:    channel,
		username:   username,
		client:     &http.Client{},
	}
}

// Send delivers the alert message to Rocket.Chat.
func (r *RocketChatNotifier) Send(subject, body string) error {
	payload := rocketChatPayload{
		Text:     fmt.Sprintf("*%s*\n%s", subject, body),
		Channel:  r.channel,
		Username: r.username,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("rocketchat: marshal payload: %w", err)
	}

	resp, err := r.client.Post(r.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("rocketchat: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("rocketchat: unexpected status %d", resp.StatusCode)
	}
	return nil
}
