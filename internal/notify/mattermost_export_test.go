package notify

import "net/http"

// SetMattermostHTTPClient replaces the HTTP client for testing.
func SetMattermostHTTPClient(n *MattermostNotifier, c *http.Client) {
	n.client = c
}
