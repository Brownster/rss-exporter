package collectors

import (
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/suite"

	"github.com/4O4-Not-F0und/rss-exporter/connectors"
	"github.com/alecthomas/kingpin/v2"
	maas "github.com/sabio-engineering-product/monitoring-maas"
)

type FeedTestSuite struct {
	suite.Suite
	Connector *connectors.MockHTTPConnector
	Exporter  *maas.Exporter
}

func (s *FeedTestSuite) SetupTest() {
	s.Connector = &connectors.MockHTTPConnector{Responses: make(map[string]string)}
	data, _ := os.ReadFile("testdata/aws_outage.rss")
	s.Connector.Responses["http://mock.aws/feed"] = string(data)

	app := kingpin.New("test", "")
	serviceCfg := maas.ServiceFeed{Name: "aws-test", URL: "http://mock.aws/feed", Interval: 300}

	e, err := maas.NewExporter(app, s.Connector,
		maas.WithScheduledScrapers(NewFeedCollector(app, serviceCfg)),
		maas.WithLabels(&maas.MockLabels{}),
	)
	s.NoError(err)
	s.Exporter = e
}

func (s *FeedTestSuite) TestAWSOutage() {
	s.Exporter.Start()

	expected := "# HELP aws_test_service_status Current service status\n" +
		"# TYPE aws_test_service_status gauge\n" +
		"aws_test_service_status{customer=\"\",service=\"aws-test\",state=\"ok\"} 0\n" +
		"aws_test_service_status{customer=\"\",service=\"aws-test\",state=\"outage\"} 1\n" +
		"aws_test_service_status{customer=\"\",service=\"aws-test\",state=\"service_issue\"} 0\n"
	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "aws-test_service_status")
	s.NoError(err)
}

func TestFeedSuite(t *testing.T) {
	suite.Run(t, new(FeedTestSuite))
}
