package collectors

import (
	"strings"

	"github.com/mmcdole/gofeed"
)

// extractServiceStatus determines the service state from a feed item.
func extractServiceStatus(item *gofeed.Item) (service string, state string, active bool) {
	upper := func(s string) string {
		return strings.ToUpper(strings.TrimSpace(s))
	}

	title := upper(item.Title)
	summary := upper(item.Description)
	content := upper(item.Content)
	combined := strings.Join([]string{title, summary, content}, " ")

	switch {
	case strings.Contains(combined, "STATUS: RESOLVED") || strings.Contains(title, "RESOLVED"):
		state = "resolved"
	case strings.Contains(combined, "MAINTENANCE") && strings.Contains(combined, "COMPLETED"):
		state = "resolved"
	case strings.Contains(combined, "OUTAGE"):
		state = "outage"
	case strings.Contains(combined, "SCHEDULED") || strings.Contains(combined, "MAINTENANCE"):
		state = "service_issue"
	case strings.Contains(combined, "INVESTIGATING"):
		state = "service_issue"
	case strings.Contains(combined, "IDENTIFIED"):
		state = "service_issue"
	case strings.Contains(combined, "MONITORING"):
		state = "service_issue"
	case strings.Contains(combined, "DEGRADED"):
		state = "service_issue"
	case strings.Contains(combined, "SERVICE ISSUE") || strings.Contains(combined, "SERVICE IMPACT"):
		state = "service_issue"
	}
	if state == "" {
		return
	}
	service = strings.TrimSpace(item.Title)
	active = state != "resolved"
	return
}
