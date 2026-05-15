package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const telegramBaseURL = "https://api.telegram.org"

// TelegramNotifier sends alert messages via the Telegram Bot API.
type TelegramNotifier struct {
	botToken string
	chatID   string
	baseURL  string
	client   *http.Client
}

type telegramPayload struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// NewTelegramNotifier creates a new TelegramNotifier.
func NewTelegramNotifier(botToken, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		baseURL:  telegramBaseURL,
		client:   &http.Client{},
	}
}

// Send delivers the alert message to the configured Telegram chat.
func (t *TelegramNotifier) Send(subject, body string) error {
	text := fmt.Sprintf("*%s*\n%s", subject, body)
	payload := telegramPayload{
		ChatID:    t.chatID,
		Text:      text,
		ParseMode: "Markdown",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", t.baseURL, t.botToken)
	resp, err := t.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("telegram: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram: unexpected status %d", resp.StatusCode)
	}
	return nil
}
