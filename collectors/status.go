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

	containsAny := func(haystack string, needles ...string) bool {
		for _, needle := range needles {
			if strings.Contains(haystack, needle) {
				return true
			}
		}
		return false
	}

	title := upper(item.Title)
	summary := upper(item.Description)
	content := upper(item.Content)
	combined := strings.Join([]string{title, summary, content}, " ")

	switch {
	case containsAny(combined, "STATUS: RESOLVED", "RESOLVED", "COMPLETED", "RESTORED"):
		state = "resolved"

	case containsAny(combined, "OUTAGE", "MAJOR INCIDENT", "SERVICE INTERRUPTION"):
		state = "outage"
	case containsAny(combined,
		"SERVICE ISSUE",
		"SERVICE IMPACT",
		"IDENTIFIED",
		"INVESTIGATING",
		"MONITORING",
		"IN PROGRESS",
		"UPDATE",
		"DELAY",
		"DEGRADED",
		"IMPAIRED",
		"MAINTENANCE",
		"SCHEDULED",
	):

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
