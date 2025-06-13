package main

import (
	"strings"

	"github.com/mmcdole/gofeed"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "rss_exporter"

var (
	serviceStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_status",
			Help:      "Current service status parsed from configured feeds.",
		},
		[]string{"service", "customer", "state"},
	)
	serviceIssueInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_issue_info",
			Help:      "Details for the currently active service issue.",
		},
		[]string{"service", "customer", "service_name", "region", "title", "link", "guid"},
	)
)

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
