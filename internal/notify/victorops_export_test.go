package notify

// SetVictorOpsBaseURL overrides the base URL for testing.
func SetVictorOpsBaseURL(n *VictorOpsNotifier, url string) {
	n.baseURL = url
}
