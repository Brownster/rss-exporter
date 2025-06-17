package maas

import (
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var metricDescriptions = map[string]*Metric{}

// WithDescription registers metric help text and label names. This mimics the
// behaviour of the real maas library but in a simplified form.
func WithDescription(_ interface{}, name, help string, labels []string) {
	metricDescriptions[name] = &Metric{Name: name, Help: help, LabelNames: labels}
}

// Connector represents a data source connection.
// It mirrors the interface provided by the real monitoring-maas library.
type Connector interface {
	Execute(query interface{}) (interface{}, error)
}

// Metric represents a single Prometheus metric returned by a scraper.
type Metric struct {
	Name        string
	Help        string
	ValueType   prometheus.ValueType
	Value       float64
	LabelNames  []string
	LabelValues []string
}

// MockLabels is a stub used for tests that require the real library's
// label provider. It carries no behaviour in this simplified version.
type MockLabels struct{}

// ServiceFeed mirrors the configuration for a single RSS feed.
type ServiceFeed struct {
	Name     string
	Provider string
	Customer string
	URL      string
	Interval int
}

// NewMetric creates a Metric instance.
func NewMetric(name string, valueType prometheus.ValueType, value float64, labels []string) Metric {
	return Metric{Name: name, ValueType: valueType, Value: value, LabelValues: labels}
}

// Scraper fetches metrics using a Connector.
type Scraper interface {
	Scrape(c Connector) ([]Metric, error)
}

// Schedule defines how often a ScheduledScraper runs.
type Schedule struct {
	Frequency time.Duration
}

// NewSchedule returns a schedule with optional modifiers.
func NewSchedule(opts ...func(*Schedule)) *Schedule {
	s := &Schedule{Frequency: time.Minute}
	for _, o := range opts {
		o(s)
	}
	return s
}

// WithFrequency sets the scraping frequency.
func WithFrequency(d time.Duration) func(*Schedule) {
	return func(s *Schedule) { s.Frequency = d }
}

// ScheduledScraper periodically executes a Scraper and stores the results.
type ScheduledScraper struct {
	Name     string
	Scraper  Scraper
	Schedule *Schedule

	mu      sync.RWMutex
	metrics []Metric
}

// NewScheduledScraper creates a ScheduledScraper instance.
func NewScheduledScraper(name string, scraper Scraper, opts ...func(*ScheduledScraper)) *ScheduledScraper {
	ss := &ScheduledScraper{Name: name, Scraper: scraper, Schedule: NewSchedule()}
	for _, o := range opts {
		o(ss)
	}
	return ss
}

// WithSchedule sets the schedule on a ScheduledScraper.
func WithSchedule(s *Schedule) func(*ScheduledScraper) {
	return func(ss *ScheduledScraper) { ss.Schedule = s }
}

// Start begins the scraping loop.
func (s *ScheduledScraper) Start(c Connector) {
	go func() {
		// Run once immediately
		s.run(c)
		ticker := time.NewTicker(s.Schedule.Frequency)
		defer ticker.Stop()
		for range ticker.C {
			s.run(c)
		}
	}()
}

func (s *ScheduledScraper) run(c Connector) {
	metrics, err := s.Scraper.Scrape(c)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.metrics = metrics
	s.mu.Unlock()
}

// Exporter collects metrics from multiple ScheduledScrapers.
type Exporter struct {
	connector Connector
	scrapers  []*ScheduledScraper
}

// NewExporter constructs an Exporter.
func NewExporter(_ interface{}, c Connector, opts ...func(*Exporter)) (*Exporter, error) {
	e := &Exporter{connector: c}
	for _, o := range opts {
		o(e)
	}
	return e, nil
}

// WithScheduledScrapers registers scrapers with the exporter.
func WithScheduledScrapers(scrapers ...*ScheduledScraper) func(*Exporter) {
	return func(e *Exporter) {
		e.scrapers = append(e.scrapers, scrapers...)
	}
}

// WithLabels is accepted for API compatibility but ignored in this stub.
func WithLabels(_ interface{}) func(*Exporter) {
	return func(e *Exporter) {}
}

// Start launches all scheduled scrapers.
func (e *Exporter) Start() {
	for _, s := range e.scrapers {
		s.Start(e.connector)
	}
}

// Serve is a no-op for this stub implementation.
func (e *Exporter) Serve() {}

// Describe sends metric descriptors to the channel.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// Descriptors are created dynamically during Collect.
}

// Collect gathers metrics from all scrapers.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	for _, s := range e.scrapers {
		s.mu.RLock()
		metrics := append([]Metric(nil), s.metrics...)
		s.mu.RUnlock()
		for _, m := range metrics {
			descInfo := metricDescriptions[m.Name]
			help := m.Help
			labels := m.LabelNames
			if descInfo != nil {
				if help == "" {
					help = descInfo.Help
				}
				if len(labels) == 0 {
					labels = descInfo.LabelNames
				}
			}
			metricName := strings.ReplaceAll(s.Name, "-", "_")
			desc := prometheus.NewDesc(prometheus.BuildFQName(metricName, "", m.Name), help, labels, nil)
			ch <- prometheus.MustNewConstMetric(desc, m.ValueType, m.Value, m.LabelValues...)
		}
	}
}
