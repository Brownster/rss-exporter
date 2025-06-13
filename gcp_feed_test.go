package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

func loadGCPFeed(t *testing.T) *gofeed.Feed {
	t.Helper()
	data, err := ioutil.ReadFile("testdata/gcp_feed.atom")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	feed, err := gofeed.NewParser().ParseString(string(data))
	if err != nil {
		t.Fatalf("parse feed: %v", err)
	}
	return feed
}

func loadGCPUpdateFeed(t *testing.T) *gofeed.Feed {
	t.Helper()
	data, err := ioutil.ReadFile("testdata/gcp_update_resolved.atom")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	feed, err := gofeed.NewParser().ParseString(string(data))
	if err != nil {
		t.Fatalf("parse feed: %v", err)
	}
	return feed
}

func loadGCPResolvedThenUpdateFeed(t *testing.T) *gofeed.Feed {
	t.Helper()
	data, err := ioutil.ReadFile("testdata/gcp_resolved_then_update.atom")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	feed, err := gofeed.NewParser().ParseString(string(data))
	if err != nil {
		t.Fatalf("parse feed: %v", err)
	}
	return feed
}

func TestExtractServiceStatus_GCPFeed(t *testing.T) {
	feed := loadGCPFeed(t)
	if len(feed.Items) == 0 {
		t.Fatal("no items in feed")
	}
	svc, state, active := extractServiceStatus(feed.Items[0])
	wantSvc := strings.TrimSpace(feed.Items[0].Title)
	if svc != wantSvc {
		t.Errorf("service got %q want %q", svc, wantSvc)
	}
	if state != "resolved" {
		t.Errorf("state got %q want resolved", state)
	}
	if active {
		t.Error("expected active false")
	}
}

func TestUpdateServiceStatus_GCPFeed(t *testing.T) {
	// serve the feed via test server
	data, err := ioutil.ReadFile("testdata/gcp_feed.atom")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	metricsMu.Lock()
	metricsData = map[string]*serviceMetrics{}
	metricsMu.Unlock()

	cfg := ServiceFeed{Name: "gcp", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	metricsMu.Lock()
	sm := metricsData["gcp"]
	metricsMu.Unlock()

	if sm.State != "ok" {
		t.Errorf("state = %v, want ok", sm.State)
	}
}

func TestExtractServiceStatus_GCPUpdateFeed(t *testing.T) {
	feed := loadGCPUpdateFeed(t)
	if len(feed.Items) == 0 {
		t.Fatal("no items in feed")
	}
	_, state, active := extractServiceStatus(feed.Items[0])
	if state != "service_issue" {
		t.Errorf("state got %q want service_issue", state)
	}
	if !active {
		t.Error("expected active true")
	}
}

func TestUpdateServiceStatus_GCPResolvedThenUpdate(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/gcp_resolved_then_update.atom")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	metricsMu.Lock()
	metricsData = map[string]*serviceMetrics{}
	metricsMu.Unlock()

	cfg := ServiceFeed{Name: "gcp", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	metricsMu.Lock()
	sm := metricsData["gcp"]
	metricsMu.Unlock()

	if sm.State != "ok" {
		t.Errorf("state = %v, want ok", sm.State)
	}
}
