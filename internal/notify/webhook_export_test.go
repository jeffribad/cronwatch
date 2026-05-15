package notify

import "net/http"

// SetWebhookHTTPClient replaces the HTTP client for testing.
func SetWebhookHTTPClient(w *WebhookNotifier, c *http.Client) {
	w.client = c
}
