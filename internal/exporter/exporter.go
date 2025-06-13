package exporter

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	defaultListenAddress = "0.0.0.0"
	defaultListenPort    = 9191
	defaultConfigFile    = "config.yml"
	defaultTimeout       = 10 * time.Second
	defaultFetchRetries  = 3
)

type Config struct {
	ListenAddress string        `yaml:"listen_address"`
	ListenPort    int           `yaml:"listen_port"`
	LogLevel      string        `yaml:"log_level"`
	Services      []ServiceFeed `yaml:"services"`
}

type ServiceFeed struct {
	Name     string `yaml:"name"`
	Provider string `yaml:"provider"`
	Customer string `yaml:"customer"`
	URL      string `yaml:"url"`
	Interval int    `yaml:"interval"`
}

var (
	AppConfig Config
	logLevels = map[string]logrus.Level{
		"trace": logrus.TraceLevel,
		"debug": logrus.DebugLevel,
		"info":  logrus.InfoLevel,
		"warn":  logrus.WarnLevel,
	}
)

func init() {
	var configFile string
	flag.StringVar(&configFile, "config", defaultConfigFile, "path to config file")
	// Skip parsing if running under "go test".
	if !strings.HasSuffix(os.Args[0], ".test") {
		flag.Parse()
	}

	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat:        time.RFC3339,
		DisableColors:          true,
		DisableLevelTruncation: true,
		ForceQuote:             true,
		FullTimestamp:          true,
	})

	var err error
	AppConfig, err = loadConfig(configFile)
	if err != nil {
		if strings.HasSuffix(os.Args[0], ".test") {
			logrus.Warnf("load config failed: %v", err)
		} else {
			logrus.Fatalf("load config failed: %v", err)
		}
	}

	prometheus.MustRegister(metricsCollector{})

	logLevel, ok := logLevels[AppConfig.LogLevel]
	if ok {
		logrus.SetLevel(logLevel)
	} else {
		logrus.Warnf("Invalid log level '%s'", AppConfig.LogLevel)
	}
}

func loadConfig(configFile string) (cfg Config, err error) {

	cfg.ListenAddress = defaultListenAddress
	cfg.ListenPort = defaultListenPort

	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("Config file '%s' not found, starting with default config", configFile)
			return
		}
		return cfg, fmt.Errorf("load config file '%s' failed: %v", configFile, err)
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("parse '%s' failed: %v", configFile, err)
	}

	for i := range cfg.Services {
		if cfg.Services[i].Interval <= 0 {
			cfg.Services[i].Interval = 300
		}
	}

	logrus.Infof("loaded config from '%s'", configFile)
	return cfg, nil
}

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

func fetchFeedWithRetry(url string, logger *logrus.Entry) (*gofeed.Feed, error) {
	backoff := time.Second
	var lastErr error
	for i := 1; i <= defaultFetchRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		feed, err := gofeed.NewParser().ParseURLWithContext(url, ctx)
		cancel()
		if err == nil {
			return feed, nil
		}
		lastErr = err
		logger.Debugf("attempt %d failed: %v", i, err)
		if i < defaultFetchRetries {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	return nil, lastErr
}
