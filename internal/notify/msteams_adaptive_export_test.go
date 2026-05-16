package notify

import "net/http"

// SetMSTeamsAdaptiveHTTPClient overrides the HTTP client for testing.
func SetMSTeamsAdaptiveHTTPClient(n *MSTeamsAdaptiveNotifier, c *http.Client) {
	n.client = c
}
