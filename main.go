package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	defaultProbePath     = "/probe"
	defaultListenAddress = "0.0.0.0"
	defaultListenPort    = 9191
	defaultConfigFile    = "config.yml"
)

type Config struct {
	ListenAddress  string        `yaml:"listen_address"`
	ListenPort     int           `yaml:"listen_port"`
	LogLevel       string        `yaml:"log_level"`
	ProbePath      string        `yaml:"probe_path"`
	DefaultTimeout int           `yaml:"default_timeout"`
	Services       []ServiceFeed `yaml:"services"`
}

type ServiceFeed struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Interval int    `yaml:"interval"`
}

var (
	appConfig Config
	enableAWS bool
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
	flag.BoolVar(&enableAWS, "enable-aws-feeds", false, "monitor default AWS service feeds")
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
	appConfig, err = loadConfig(configFile)
	if err != nil {
		if strings.HasSuffix(os.Args[0], ".test") {
			logrus.Warnf("load config failed: %v", err)
		} else {
			logrus.Fatalf("load config failed: %v", err)
		}
	}

	if enableAWS {
		appConfig.Services = append(appConfig.Services, defaultAWSServiceFeeds()...)
	}

	prometheus.MustRegister(serviceStatusGauge)

	logLevel, ok := logLevels[appConfig.LogLevel]
	if ok {
		logrus.SetLevel(logLevel)
	} else {
		logrus.Warnf("Invalid log level '%s'", appConfig.LogLevel)
	}
}

func loadConfig(configFile string) (cfg Config, err error) {

	cfg.ListenAddress = defaultListenAddress
	cfg.ListenPort = defaultListenPort
	cfg.ProbePath = defaultProbePath
	cfg.DefaultTimeout = int(defaultTimeout.Seconds())

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

	if cfg.ProbePath == "" {
		cfg.ProbePath = defaultProbePath
	} else if cfg.ProbePath[0] != '/' {
		cfg.ProbePath = "/" + cfg.ProbePath
	}

	logrus.Infof("loaded config from '%s'", configFile)
	return cfg, nil
}

func main() {
	for _, svc := range appConfig.Services {
		go monitorService(svc)
	}

	http.HandleFunc(appConfig.ProbePath, probeHandler)
	http.HandleFunc("/", landingPageHandler)
	http.Handle("/metrics", promhttp.Handler())

	listenOn := fmt.Sprintf("%s:%d", appConfig.ListenAddress, appConfig.ListenPort)
	logrus.Infof("Listening at: %s", listenOn)
	logrus.Fatalln(http.ListenAndServe(listenOn, nil))
}

func landingPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	landingPageHTML := fmt.Sprintf(`<html>
                        <head><title>RSS Exporter</title></head>
                        <body>
                        <h1>RSS Exporter</h1>
                        <p>Probe Example: <code>%s?target=https://example.com/files/sample-rss-2.xml&timeout=10</code></p>
                        </body>
                        </html>`, appConfig.ProbePath)
	w.Write([]byte(landingPageHTML))
}

func monitorService(cfg ServiceFeed) {
	logger := logrus.WithField("service", cfg.Name)
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	for {
		updateServiceStatus(cfg, logger)
		<-ticker.C
	}
}

func updateServiceStatus(cfg ServiceFeed, logger *logrus.Entry) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(appConfig.DefaultTimeout)*time.Second)
	defer cancel()

	feed, err := gofeed.NewParser().ParseURLWithContext(cfg.URL, ctx)
	if err != nil {
		logger.Warnf("fetch feed failed: %v", err)
		return
	}

	state := "ok"
	if len(feed.Items) > 0 {
		_, st, active := extractServiceStatus(feed.Items[0])
		if active {
			state = st
		}
	}

	for _, s := range []string{"ok", "service_issue", "outage"} {
		val := 0.0
		if s == state {
			val = 1
		}
		serviceStatusGauge.WithLabelValues(cfg.Name, s).Set(val)
	}
}
