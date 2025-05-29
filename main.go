package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

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
	ListenAddress  string `yaml:"listen_address"`
	ListenPort     int    `yaml:"listen_port"`
	LogLevel       string `yaml:"log_level"`
	ProbePath      string `yaml:"probe_path"`
	DefaultTimeout int    `yaml:"default_timeout"`
}

var (
	appConfig Config
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
	flag.Parse()

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
		logrus.Fatalf("load config failed: %v", err)
	}

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

	if cfg.ProbePath == "" {
		cfg.ProbePath = defaultProbePath
	} else if cfg.ProbePath[0] != '/' {
		cfg.ProbePath = "/" + cfg.ProbePath
	}

	logrus.Infof("loaded config from '%s'", configFile)
	return cfg, nil
}

func main() {
	http.HandleFunc(appConfig.ProbePath, probeHandler)
	http.HandleFunc("/", landingPageHandler)

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
