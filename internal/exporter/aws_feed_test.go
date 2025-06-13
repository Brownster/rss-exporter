package exporter

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func loadAWSFeed(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/aws_feed.rss")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	return data
}

func loadAWSAthenaIssueFeed(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/aws_athena_us_west_2_issue.rss")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	return data
}

func loadAWSMultiItemFeed(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/aws_multi_item.rss")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	return data
}

func TestUpdateServiceStatus_AWSFeed(t *testing.T) {
	data := loadAWSFeed(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	metricsMu.Lock()
	metricsData = map[string]*serviceMetrics{}
	metricsMu.Unlock()

	cfg := ServiceFeed{Name: "aws", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	metricsMu.Lock()
	sm := metricsData["aws"]
	metricsMu.Unlock()

	if sm.State != "ok" {
		t.Errorf("state = %v, want ok", sm.State)
	}
}

func TestUpdateServiceStatus_AWSAthenaIssueFeed(t *testing.T) {
	data := loadAWSAthenaIssueFeed(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	metricsMu.Lock()
	metricsData = map[string]*serviceMetrics{}
	metricsMu.Unlock()

	cfg := ServiceFeed{Name: "aws-athena", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	metricsMu.Lock()
	sm := metricsData["aws-athena"]
	metricsMu.Unlock()

	if sm.State != "service_issue" {
		t.Errorf("state = %v, want service_issue", sm.State)
	}
}

func TestServiceIssueInfoMetric(t *testing.T) {
	data := loadAWSAthenaIssueFeed(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	metricsMu.Lock()
	metricsData = map[string]*serviceMetrics{}
	metricsMu.Unlock()

	cfg := ServiceFeed{Name: "aws-athena", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	metricsMu.Lock()
	sm := metricsData["aws-athena"]
	metricsMu.Unlock()

	if sm.Issue == nil {
		t.Fatal("expected issue info")
	}
	if sm.Issue.ServiceName != "athena" || sm.Issue.Region != "us-west-2" {
		t.Errorf("issue info mismatch: %+v", sm.Issue)
	}
}

func TestUpdateServiceStatus_AWSMultiItemFeed(t *testing.T) {
	data := loadAWSMultiItemFeed(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	metricsMu.Lock()
	metricsData = map[string]*serviceMetrics{}
	metricsMu.Unlock()

	cfg := ServiceFeed{Name: "aws-multi", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	metricsMu.Lock()
	sm := metricsData["aws-multi"]
	metricsMu.Unlock()

	if sm.State != "ok" {
		t.Errorf("state = %v, want ok", sm.State)
	}
}

func loadAWSOutageFeed(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/aws_outage.rss")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	return data
}

func TestUpdateServiceStatus_AWSOutageFeed(t *testing.T) {
	data := loadAWSOutageFeed(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	metricsMu.Lock()
	metricsData = map[string]*serviceMetrics{}
	metricsMu.Unlock()

	cfg := ServiceFeed{Name: "aws-ec2", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	metricsMu.Lock()
	sm := metricsData["aws-ec2"]
	metricsMu.Unlock()

	if sm.State != "outage" {
		t.Errorf("state = %v, want outage", sm.State)
	}
	if sm.Issue == nil || sm.Issue.ServiceName != "ec2" || sm.Issue.Region != "us-west-2" {
		t.Errorf("issue info mismatch: %+v", sm.Issue)
	}
}
