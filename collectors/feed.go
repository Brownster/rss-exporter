package collectors

import (
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/mmcdole/gofeed"
	"github.com/prometheus/client_golang/prometheus"
	maas "github.com/sabio-engineering-product/monitoring-maas"

	"github.com/4O4-Not-F0und/rss-exporter/connectors"
)

// NewFeedCollector creates a scheduled scraper for a single RSS feed.
func NewFeedCollector(app *kingpin.Application, serviceConfig maas.ServiceFeed) *maas.ScheduledScraper {
	maas.WithDescription(app, "service_status", "Current service status", []string{"service", "customer", "state"})
	maas.WithDescription(app, "service_issue_info", "Details for active service issues", []string{"service", "customer", "service_name", "region", "title", "link", "guid"})

	return maas.NewScheduledScraper(
		serviceConfig.Name,
		NewFeedScraper(serviceConfig),
		maas.WithSchedule(maas.NewSchedule(
			maas.WithFrequency(time.Duration(serviceConfig.Interval)*time.Second),
		)),
	)
}

// FeedScraper holds configuration for scraping a feed.
type FeedScraper struct {
	Config maas.ServiceFeed
	Parser Scraper
}

// NewFeedScraper returns a new FeedScraper instance.
func NewFeedScraper(cfg maas.ServiceFeed) *FeedScraper {
	return &FeedScraper{
		Config: cfg,
		Parser: ScraperForService(cfg.Provider, cfg.Name),
	}
}

// Scrape fetches the feed and converts status into metrics.
func (s *FeedScraper) Scrape(c maas.Connector) ([]maas.Metric, error) {
	metrics := []maas.Metric{}

	feed, err := c.Execute(connectors.HTTPQuery{URL: s.Config.URL})
	if err != nil {
		return nil, err
	}

	fp := feed.(*gofeed.Feed)

	state := "ok"
	var activeItem *gofeed.Item
	scraper := s.Parser
	var svcName, region string
	seen := make(map[string]struct{})
	for _, item := range fp.Items {
		key := scraper.IncidentKey(item)
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		_, st, active := extractServiceStatus(item)
		if st == "resolved" {
			state = "ok"
			activeItem = nil
			svcName, region = scraper.ServiceInfo(item)
			break
		}
		if active {
			state = st
			activeItem = item
			svcName, region = scraper.ServiceInfo(item)
			break
		}
	}

	for _, st := range []string{"ok", "service_issue", "outage"} {
		val := 0.0
		if state == st {
			val = 1.0
		}
		metrics = append(metrics, maas.NewMetric("service_status", prometheus.GaugeValue, val, []string{s.Config.Name, s.Config.Customer, st}))
	}

	if activeItem != nil {
		if svcName == "" && region == "" {
			svcName, region = scraper.ServiceInfo(activeItem)
		}
		metrics = append(metrics, maas.NewMetric("service_issue_info", prometheus.GaugeValue, 1, []string{s.Config.Name, s.Config.Customer, svcName, region, strings.TrimSpace(activeItem.Title), activeItem.Link, activeItem.GUID}))
	}

	return metrics, nil
}
