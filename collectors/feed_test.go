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
	s.Exporter = nil
}

func (s *FeedTestSuite) setupExporter(feedPath, url, name, provider string) {
	data, err := os.ReadFile(feedPath)
	s.Require().NoError(err)
	s.Connector.Responses[url] = string(data)

	app := kingpin.New("test", "")
	cfg := maas.ServiceFeed{Name: name, URL: url, Provider: provider, Interval: 300}

	e, err := maas.NewExporter(app, s.Connector,
		maas.WithScheduledScrapers(NewFeedCollector(app, cfg)),
		maas.WithLabels(&maas.MockLabels{}),
	)
	s.Require().NoError(err)
	s.Exporter = e
}

func (s *FeedTestSuite) TestAWSOutage() {
	s.setupExporter("testdata/aws_outage.rss", "http://mock.aws/feed", "aws-test", "aws")
	s.Exporter.Start()

	expected := "# HELP aws_test_service_status Current service status\n" +
		"# TYPE aws_test_service_status gauge\n" +
		"aws_test_service_status{customer=\"\",service=\"aws-test\",state=\"ok\"} 0\n" +
		"aws_test_service_status{customer=\"\",service=\"aws-test\",state=\"outage\"} 1\n" +
		"aws_test_service_status{customer=\"\",service=\"aws-test\",state=\"service_issue\"} 0\n"
	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "aws-test_service_status")
	s.NoError(err)
}

func (s *FeedTestSuite) TestAzureServiceIssue() {
	s.setupExporter("testdata/azure_issue.rss", "http://mock.azure/feed", "azure-test", "azure")
	s.Exporter.Start()

	expected := "# HELP azure_test_service_status Current service status\n" +
		"# TYPE azure_test_service_status gauge\n" +
		"azure_test_service_status{customer=\"\",service=\"azure-test\",state=\"ok\"} 0\n" +
		"azure_test_service_status{customer=\"\",service=\"azure-test\",state=\"outage\"} 0\n" +
		"azure_test_service_status{customer=\"\",service=\"azure-test\",state=\"service_issue\"} 1\n"
	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "azure-test_service_status")
	s.NoError(err)
}

func (s *FeedTestSuite) TestOpenAIResolved() {
	s.setupExporter("testdata/openai_resolved.atom", "http://mock.openai/feed", "openai-test", "")
	s.Exporter.Start()

	expected := "# HELP openai_test_service_status Current service status\n" +
		"# TYPE openai_test_service_status gauge\n" +
		"openai_test_service_status{customer=\"\",service=\"openai-test\",state=\"ok\"} 1\n" +
		"openai_test_service_status{customer=\"\",service=\"openai-test\",state=\"outage\"} 0\n" +
		"openai_test_service_status{customer=\"\",service=\"openai-test\",state=\"service_issue\"} 0\n"
	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "openai-test_service_status")
	s.NoError(err)
}

func (s *FeedTestSuite) TestGenesysIdentifiedIncident() {
	s.setupExporter("testdata/genesys_feed.atom", "http://mock.genesys/feed", "genesys-test", "genesyscloud")
	s.Exporter.Start()

	expected := "# HELP genesys_test_service_status Current service status\n" +
		"# TYPE genesys_test_service_status gauge\n" +
		"genesys_test_service_status{customer=\"\",service=\"genesys-test\",state=\"ok\"} 0\n" +
		"genesys_test_service_status{customer=\"\",service=\"genesys-test\",state=\"outage\"} 0\n" +
		"genesys_test_service_status{customer=\"\",service=\"genesys-test\",state=\"service_issue\"} 1\n"
	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "genesys-test_service_status")
	s.NoError(err)
}

func TestFeedSuite(t *testing.T) {
	suite.Run(t, new(FeedTestSuite))
}
