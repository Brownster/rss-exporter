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

func (s *FeedTestSuite) TestOpenAIServiceIssue() {
	s.setupExporter("testdata/openai_monitoring.rss", "http://mock.openai/feed", "openai-test", "")
	s.Exporter.Start()

	expected := "# HELP openai_test_service_status Current service status\n" +
		"# TYPE openai_test_service_status gauge\n" +
		"openai_test_service_status{customer=\"\",service=\"openai-test\",state=\"ok\"} 0\n" +
		"openai_test_service_status{customer=\"\",service=\"openai-test\",state=\"outage\"} 0\n" +
		"openai_test_service_status{customer=\"\",service=\"openai-test\",state=\"service_issue\"} 1\n"
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

	expectedInfo := "# HELP genesys_test_service_issue_info Details for active service issues\n" +
		"# TYPE genesys_test_service_issue_info gauge\n" +
		"genesys_test_service_issue_info{customer=\"\",guid=\"tag:status.mypurecloud.com,2005:Incident/26815638\",link=\"https://status.mypurecloud.com/incidents/nj3rdcjrgh02\",region=\"multiple-regions\",service=\"genesys-test\",service_name=\"multiple-services\",title=\"Americas (US East) Incident\"} 1\n"
	err = testutil.CollectAndCompare(s.Exporter, strings.NewReader(expectedInfo), "genesys-test_service_issue_info")
	s.NoError(err)
}

func (s *FeedTestSuite) TestTwilioScheduledMaintenance() {
	s.setupExporter("testdata/twilio_scheduled.rss", "http://mock.twilio/feed", "twilio-test", "twilio")
	s.Exporter.Start()

	expectedStatus := "# HELP twilio_test_service_status Current service status\n" +
		"# TYPE twilio_test_service_status gauge\n" +
		"twilio_test_service_status{customer=\"\",service=\"twilio-test\",state=\"ok\"} 0\n" +
		"twilio_test_service_status{customer=\"\",service=\"twilio-test\",state=\"outage\"} 0\n" +
		"twilio_test_service_status{customer=\"\",service=\"twilio-test\",state=\"service_issue\"} 1\n"
	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expectedStatus), "twilio-test_service_status")
	s.NoError(err)

	expectedInfo := "# HELP twilio_test_service_issue_info Details for active service issues\n" +
		"# TYPE twilio_test_service_issue_info gauge\n" +
		"twilio_test_service_issue_info{customer=\"\",guid=\"https://status.twilio.com/incidents/2v98q17wsch8\",link=\"https://status.twilio.com/incidents/2v98q17wsch8\",region=\"United States and Canada\",service=\"twilio-test\",service_name=\"SMS and MMS\",title=\"United States and Canada Twilio SMS and MMS Maintenance\"} 1\n"
	err = testutil.CollectAndCompare(s.Exporter, strings.NewReader(expectedInfo), "twilio-test_service_issue_info")

	s.NoError(err)
}

func TestFeedSuite(t *testing.T) {
	suite.Run(t, new(FeedTestSuite))
}
