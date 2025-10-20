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
		state = "service_issue"
	}
	if state == "" {
		return
	}
	service = strings.TrimSpace(item.Title)
	active = state != "resolved"
	return
}
