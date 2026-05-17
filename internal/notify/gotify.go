package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// GotifyNotifier sends alerts to a self-hosted Gotify server.
type GotifyNotifier struct {
	baseURL  string
	token    string
	priority int
	client   *http.Client
}

type gotifyPayload struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

// NewGotifyNotifier creates a GotifyNotifier.
// baseURL is the root URL of the Gotify server (e.g. "https://gotify.example.com").
// token is the Gotify application token.
// priority sets the message priority (0–10); 0 means default (5).
func NewGotifyNotifier(baseURL, token string, priority int) *GotifyNotifier {
	if priority <= 0 {
		priority = 5
	}
	return &GotifyNotifier{
		baseURL:  baseURL,
		token:    token,
		priority: priority,
		client:   &http.Client{},
	}
}

// Send delivers the alert to Gotify.
func (g *GotifyNotifier) Send(subject, body string) error {
	payload := gotifyPayload{
		Title:    subject,
		Message:  body,
		Priority: g.priority,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("gotify: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/message?token=%s", g.baseURL, g.token)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("gotify: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("gotify: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gotify: unexpected status %d", resp.StatusCode)
	}
	return nil
}
