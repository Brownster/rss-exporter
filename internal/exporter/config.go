package exporter

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	defaultListenAddress = "0.0.0.0"
	defaultListenPort    = 9191
	defaultConfigFile    = "config.yml"
)

// Config defines exporter settings loaded from YAML configuration.
type Config struct {
	ListenAddress string        `yaml:"listen_address"`
	ListenPort    int           `yaml:"listen_port"`
	LogLevel      string        `yaml:"log_level"`
	Services      []ServiceFeed `yaml:"services"`
}

// ServiceFeed describes a service RSS/Atom feed to monitor.
type ServiceFeed struct {
	Name     string `yaml:"name"`
	Provider string `yaml:"provider"`
	Customer string `yaml:"customer"`
	URL      string `yaml:"url"`
	Interval int    `yaml:"interval"`
}

var (
	// AppConfig holds the loaded configuration.
	AppConfig Config
	// logLevels maps string log levels to logrus constants.
	logLevels = map[string]logrus.Level{
		"trace": logrus.TraceLevel,
		"debug": logrus.DebugLevel,
		"info":  logrus.InfoLevel,
		"warn":  logrus.WarnLevel,
	}
)

// initConfig parses command line flags and loads configuration from file.
func initConfig() {
	var configFile string
	flag.StringVar(&configFile, "config", defaultConfigFile, "path to config file")
	// Skip parsing if running under "go test".
	if !strings.HasSuffix(os.Args[0], ".test") {
		flag.Parse()
	}

	var err error
	AppConfig, err = loadConfig(configFile)
	if err != nil {
		if strings.HasSuffix(os.Args[0], ".test") {
			logrus.Warnf("load config failed: %v", err)
		} else {
			logrus.Fatalf("load config failed: %v", err)
		}
	}

	logLevel, ok := logLevels[AppConfig.LogLevel]
	if ok {
		logrus.SetLevel(logLevel)
	} else {
		logrus.Warnf("Invalid log level '%s'", AppConfig.LogLevel)
	}
}

// loadConfig reads configuration from the provided YAML file.
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
