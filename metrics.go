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
		[]string{"service", "state"},
	)
	serviceIssueInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_issue_info",
			Help:      "Details for the currently active service issue.",
		},
		[]string{"service", "title", "link", "guid"},
	)
)

func extractServiceStatus(item *gofeed.Item) (service string, state string, active bool) {
	text := strings.ToUpper(strings.TrimSpace(item.Title))
	switch {
	case strings.Contains(text, "RESOLVED"):
		state = "resolved"
	case strings.Contains(text, "OUTAGE"):
		state = "outage"
	case strings.Contains(text, "SERVICE ISSUE"), strings.Contains(text, "SERVICE IMPACT"):
		state = "service_issue"
	}
	if state == "" {
		return
	}
	service = strings.TrimSpace(item.Title)
	active = state != "resolved"
	return
}
