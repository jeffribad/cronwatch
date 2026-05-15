package notify

import "net/http"

// SetTelegramHTTPClient replaces the HTTP client for testing.
func SetTelegramHTTPClient(n *TelegramNotifier, c *http.Client) {
	n.client = c
}

// SetTelegramBaseURL overrides the API base URL for testing.
func SetTelegramBaseURL(n *TelegramNotifier, url string) {
	n.baseURL = url
}
