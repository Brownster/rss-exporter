package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sirupsen/logrus"
)

func loadOpenAIResolvedFeed(t *testing.T) *gofeed.Feed {
	t.Helper()
	data, err := ioutil.ReadFile("testdata/openai_resolved.atom")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	feed, err := gofeed.NewParser().ParseString(string(data))
	if err != nil {
		t.Fatalf("parse feed: %v", err)
	}
	return feed
}

func TestExtractServiceStatus_OpenAIResolved(t *testing.T) {
	feed := loadOpenAIResolvedFeed(t)
	if len(feed.Items) == 0 {
		t.Fatal("no items in feed")
	}
	_, state, active := extractServiceStatus(feed.Items[0])
	if state != "resolved" {
		t.Errorf("state got %q want resolved", state)
	}
	if active {
		t.Error("expected active false")
	}
}

func TestUpdateServiceStatus_OpenAIResolved(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/openai_resolved.atom")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	serviceStatusGauge.Reset()

	cfg := ServiceFeed{Name: "openai", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("openai", "openai", "ok")); val != 1 {
		t.Errorf("ok gauge = %v, want 1", val)
	}
	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("openai", "openai", "service_issue")); val != 0 {
		t.Errorf("service_issue gauge = %v, want 0", val)
	}
	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("openai", "openai", "outage")); val != 0 {
		t.Errorf("outage gauge = %v, want 0", val)
	}
}
