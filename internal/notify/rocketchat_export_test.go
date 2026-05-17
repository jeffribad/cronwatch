package notify

import "net/http"

// SetRocketChatHTTPClient replaces the HTTP client for testing.
func SetRocketChatHTTPClient(n *RocketChatNotifier, c *http.Client) {
	n.client = c
}
