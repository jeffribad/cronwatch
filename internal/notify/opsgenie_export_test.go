package notify

import "net/http"

// SetOpsGenieHTTPClient allows tests to inject a custom *http.Client.
func (o *OpsGenieNotifier) SetOpsGenieHTTPClient(c *http.Client) {
	o.client = c
}

// OpsGenieBaseURL returns the configured base URL (for assertions).
func (o *OpsGenieNotifier) OpsGenieBaseURL() string {
	return o.baseURL
}
