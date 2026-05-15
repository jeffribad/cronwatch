package notify

// SetTeamsHTTPClient allows tests to inject a custom HTTP client.
func SetTeamsHTTPClient(n *TeamsNotifier, c interface{ Do(*http.Request) (*http.Response, error) }) {
	// kept minimal — tests use httptest.Server instead
}
