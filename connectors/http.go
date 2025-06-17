package connectors

import (
	connector "github.com/4O4-Not-F0und/rss-exporter/internal/connectors"
	"github.com/sirupsen/logrus"
)

// HTTPConnector implements maas.Connector for fetching RSS feeds.
type HTTPConnector struct {
	Logger *logrus.Entry
}

// NewHTTPConnector creates a new HTTP connector.
func NewHTTPConnector() *HTTPConnector {
	return &HTTPConnector{
		Logger: logrus.WithField("component", "http_connector"),
	}
}

// The query will be the URL string.
type HTTPQuery struct {
	URL string
}

// Execute fetches the RSS feed.
func (c *HTTPConnector) Execute(query interface{}) (interface{}, error) {
	httpQuery := query.(HTTPQuery)
	// Reuse existing FetchFeedWithRetry logic
	return connector.FetchFeedWithRetry(httpQuery.URL, c.Logger)
}
