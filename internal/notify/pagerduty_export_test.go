package notify

// SetPagerDutyEndpoint overrides the PagerDuty endpoint for testing.
func (p *PagerDutyNotifier) SetPagerDutyEndpoint(url string) {
	p.endpoint = url
}
