package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sirupsen/logrus"
)

func loadAWSFeed(t *testing.T) []byte {
	t.Helper()
	data, err := ioutil.ReadFile("testdata/aws_feed.rss")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	return data
}

func loadAWSAthenaIssueFeed(t *testing.T) []byte {
	t.Helper()
	data, err := ioutil.ReadFile("testdata/aws_athena_us_west_2_issue.rss")
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

	serviceStatusGauge.Reset()

	cfg := ServiceFeed{Name: "aws", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("aws", "ok")); val != 1 {
		t.Errorf("ok gauge = %v, want 1", val)
	}
	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("aws", "service_issue")); val != 0 {
		t.Errorf("service_issue gauge = %v, want 0", val)
	}
	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("aws", "outage")); val != 0 {
		t.Errorf("outage gauge = %v, want 0", val)
	}
}

func TestUpdateServiceStatus_AWSAthenaIssueFeed(t *testing.T) {
	data := loadAWSAthenaIssueFeed(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	serviceStatusGauge.Reset()

	cfg := ServiceFeed{Name: "aws-athena", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("aws-athena", "ok")); val != 0 {
		t.Errorf("ok gauge = %v, want 0", val)
	}
	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("aws-athena", "service_issue")); val != 1 {
		t.Errorf("service_issue gauge = %v, want 1", val)
	}
	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("aws-athena", "outage")); val != 0 {
		t.Errorf("outage gauge = %v, want 0", val)
	}
}
