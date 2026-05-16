package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MSTeamsAdaptiveNotifier sends alerts via MS Teams Incoming Webhook using Adaptive Cards.
type MSTeamsAdaptiveNotifier struct {
	webhookURL string
	client     *http.Client
}

type adaptiveCardPayload struct {
	Type        string              `json:"type"`
	Attachments []adaptiveAttachment `json:"attachments"`
}

type adaptiveAttachment struct {
	ContentType string          `json:"contentType"`
	Content     adaptiveContent `json:"content"`
}

type adaptiveContent struct {
	Schema  string           `json:"$schema"`
	Type    string           `json:"type"`
	Version string           `json:"version"`
	Body    []adaptiveElement `json:"body"`
}

type adaptiveElement struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	Weight string `json:"weight,omitempty"`
	Size   string `json:"size,omitempty"`
	Wrap   bool   `json:"wrap,omitempty"`
}

// NewMSTeamsAdaptiveNotifier creates a new MSTeamsAdaptiveNotifier.
func NewMSTeamsAdaptiveNotifier(webhookURL string) *MSTeamsAdaptiveNotifier {
	return &MSTeamsAdaptiveNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Send delivers an alert message via MS Teams Adaptive Card webhook.
func (n *MSTeamsAdaptiveNotifier) Send(subject, body string) error {
	payload := adaptiveCardPayload{
		Type: "message",
		Attachments: []adaptiveAttachment{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content: adaptiveContent{
					Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
					Type:    "AdaptiveCard",
					Version: "1.4",
					Body: []adaptiveElement{
						{Type: "TextBlock", Text: subject, Weight: "Bolder", Size: "Medium"},
						{Type: "TextBlock", Text: body, Wrap: true},
					},
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("msteams_adaptive: marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("msteams_adaptive: post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("msteams_adaptive: unexpected status %d", resp.StatusCode)
	}
	return nil
}
