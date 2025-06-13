package exporter

import (
	"context"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

// monitorService periodically fetches the configured feed and updates metrics.
func monitorService(ctx context.Context, cfg ServiceFeed) {
	logger := logrus.WithField("service", cfg.Name)
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	for {
		updateServiceStatus(cfg, logger)
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

// updateServiceStatus fetches the feed once and records status information.
func updateServiceStatus(cfg ServiceFeed, logger *logrus.Entry) {
	feed, err := fetchFeedWithRetry(cfg.URL, logger)
	if err != nil {
		logger.Warnf("fetch feed failed: %v", err)
		metricsMu.Lock()
		sm, ok := metricsData[cfg.Name]
		if !ok {
			sm = &serviceMetrics{Customer: cfg.Customer}
			metricsData[cfg.Name] = sm
		}
		sm.FetchErrors++
		metricsMu.Unlock()
		return
	}

	// reset error counter on success
	metricsMu.Lock()
	if sm, ok := metricsData[cfg.Name]; ok {
		sm.FetchErrors = 0
	}
	metricsMu.Unlock()

	state := "ok"
	var activeItem *gofeed.Item
	parser := parserForService(cfg.Provider, cfg.Name)
	var svcName, region string
	seen := make(map[string]struct{})
	for _, item := range feed.Items {
		key := parser.IncidentKey(item)
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		_, st, active := extractServiceStatus(item)
		if st == "resolved" {
			// issue has been resolved; ignore older items
			state = "ok"
			activeItem = nil
			svcName, region = parser.ServiceInfo(item)
			break
		}
		if active {
			state = st
			activeItem = item
			svcName, region = parser.ServiceInfo(item)
			break
		}
	}

	var info *issueInfo
	if activeItem != nil {
		if svcName == "" && region == "" {
			svcName, region = parser.ServiceInfo(activeItem)
		}
		info = &issueInfo{
			ServiceName: svcName,
			Region:      region,
			Title:       strings.TrimSpace(activeItem.Title),
			Link:        activeItem.Link,
			GUID:        activeItem.GUID,
		}
	}

	metricsMu.Lock()
	metricsData[cfg.Name] = &serviceMetrics{
		Customer: cfg.Customer,
		State:    state,
		Issue:    info,
	}
	metricsMu.Unlock()
}
