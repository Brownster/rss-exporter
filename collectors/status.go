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
	case strings.Contains(combined, "OUTAGE"):
		state = "outage"
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
