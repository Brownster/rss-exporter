package connectors

import (
	"fmt"
	"strings"

	"github.com/mmcdole/gofeed"
)

// MockHTTPConnector simulates HTTP requests for testing.
type MockHTTPConnector struct {
	Responses map[string]string // Map URL to response content
}

// Execute returns a parsed feed from the mock responses.
func (c *MockHTTPConnector) Execute(query interface{}) (interface{}, error) {
	httpQuery := query.(HTTPQuery)
	content, ok := c.Responses[httpQuery.URL]
	if !ok {
		return nil, fmt.Errorf("no mock response for URL: %s", httpQuery.URL)
	}
	return gofeed.NewParser().Parse(strings.NewReader(content))
}
