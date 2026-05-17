package notify

import "net/http"

// SetGotifyHTTPClient replaces the HTTP client for testing.
func SetGotifyHTTPClient(n *GotifyNotifier, c *http.Client) {
	n.client = c
}
