package collectors

import (
	"os"

	"github.com/alecthomas/kingpin/v2"
	maas "github.com/sabio-engineering-product/monitoring-maas"
	"gopkg.in/yaml.v3"
)

// Config defines exporter settings loaded from YAML configuration.
type Config struct {
	Services []maas.ServiceFeed `yaml:"services"`
}

// NewRssExporter constructs a maas exporter with feed scrapers based on config.
func NewRssExporter(c maas.Connector, options ...func(*maas.Exporter)) (*maas.Exporter, error) {
	app := kingpin.New("rss_exporter", "Exporter for RSS/Atom status feeds.").DefaultEnvars()

	configFile := app.Flag("config.file", "RSS exporter configuration file.").Default("config.yml").String()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	yamlFile, err := os.ReadFile(*configFile)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, err
	}

	scrapers := []*maas.ScheduledScraper{}
	for _, svc := range cfg.Services {
		if svc.Interval <= 0 {
			svc.Interval = 300
		}
		scrapers = append(scrapers, NewFeedCollector(app, svc))
	}

	options = append(options, maas.WithScheduledScrapers(scrapers...))
	return maas.NewExporter(app, c, options...)
}
