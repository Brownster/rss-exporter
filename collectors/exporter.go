package collectors

import (
	"github.com/alecthomas/kingpin/v2"
	maas "github.com/sabio-engineering-product/monitoring-maas"
)

// NewRssExporter constructs a maas exporter with feed scrapers based on config.
func NewRssExporter(c maas.Connector, options ...func(*maas.Exporter)) (*maas.Exporter, error) {
	app := kingpin.New("rss_exporter", "Exporter for RSS/Atom status feeds.").DefaultEnvars()

	scrapers := []*maas.ScheduledScraper{}
	for _, svc := range AppConfig.Services {
		scrapers = append(scrapers, NewFeedCollector(app, maas.ServiceFeed{
			Name:     svc.Name,
			Provider: svc.Provider,
			Customer: svc.Customer,
			URL:      svc.URL,
			Interval: svc.Interval,
		}))
	}

	options = append(options, maas.WithScheduledScrapers(scrapers...))
	return maas.NewExporter(app, c, options...)
}
