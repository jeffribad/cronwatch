package notify

import "net/http"

// SetDiscordHTTPClient replaces the HTTP client used by DiscordNotifier.
// For testing only.
func SetDiscordHTTPClient(d *DiscordNotifier, c *http.Client) {
	d.client = c
}
