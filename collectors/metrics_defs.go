package collectors

import (
	"strings"
	"sync"

	"github.com/mmcdole/gofeed"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "rss_exporter"

type issueInfo struct {
	ServiceName string
	Region      string
	Title       string
	Link        string
	GUID        string
}

type serviceMetrics struct {
	Customer    string
	State       string
	Issue       *issueInfo
	FetchErrors int
}

var (
	metricsMu   sync.Mutex
	metricsData = map[string]*serviceMetrics{}
)

type metricsCollector struct{}

func (metricsCollector) Describe(ch chan<- *prometheus.Desc) {}

func (metricsCollector) Collect(ch chan<- prometheus.Metric) {
	metricsMu.Lock()
	defer metricsMu.Unlock()

	for svc, sm := range metricsData {
		for _, s := range []string{"ok", "service_issue", "outage"} {
			val := 0.0
			if sm.State == s {
				val = 1
			}
			var (
				labels []string
				values []string
			)
			if sm.Customer != "" {
				labels = []string{"service", "customer", "state"}
				values = []string{svc, sm.Customer, s}
			} else {
				labels = []string{"service", "state"}
				values = []string{svc, s}
			}
			desc := prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", "service_status"),
				"Current service status parsed from configured feeds.",
				labels, nil,
			)
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, val, values...)
		}

		if sm.Issue != nil {
			var (
				labels []string
				values []string
			)
			if sm.Customer != "" {
				labels = []string{"service", "customer", "service_name", "region", "title", "link", "guid"}
				values = []string{svc, sm.Customer, sm.Issue.ServiceName, sm.Issue.Region, sm.Issue.Title, sm.Issue.Link, sm.Issue.GUID}
			} else {
				labels = []string{"service", "service_name", "region", "title", "link", "guid"}
				values = []string{svc, sm.Issue.ServiceName, sm.Issue.Region, sm.Issue.Title, sm.Issue.Link, sm.Issue.GUID}
			}
			desc := prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", "service_issue_info"),
				"Details for the currently active service issue.",
				labels, nil,
			)
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1, values...)
		}

		{
			var (
				labels []string
				values []string
			)
			if sm.Customer != "" {
				labels = []string{"service", "customer"}
				values = []string{svc, sm.Customer}
			} else {
				labels = []string{"service"}
				values = []string{svc}
			}
			desc := prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "", "fetch_errors_total"),
				"Total number of errors fetching a service feed.",
				labels, nil,
			)
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(sm.FetchErrors), values...)
		}
	}
}

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
