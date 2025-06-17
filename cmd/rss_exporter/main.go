package main

import (
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"

	"github.com/4O4-Not-F0und/rss-exporter/collectors"
	"github.com/4O4-Not-F0und/rss-exporter/connectors"
)

func main() {
	// Instantiate the exporter with the new HTTP connector
	e, err := collectors.NewRssExporter(connectors.NewHTTPConnector())
	if err != nil {
		logrus.Fatal(err)
	}

	e.Start()
	e.Serve()

	sentry.Flush(5 * time.Second)
}
