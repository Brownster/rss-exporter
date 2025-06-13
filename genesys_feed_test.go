package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sirupsen/logrus"
)

func loadGenesysFeed(t *testing.T) []byte {
	t.Helper()
	data, err := ioutil.ReadFile("testdata/genesys_feed.atom")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	return data
}

func TestUpdateServiceStatus_GenesysFeed(t *testing.T) {
	data := loadGenesysFeed(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	serviceStatusGauge.Reset()

	cfg := ServiceFeed{Name: "genesys-cloud", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("genesys-cloud", "ok")); val != 1 {
		t.Errorf("ok gauge = %v, want 1", val)
	}
	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("genesys-cloud", "service_issue")); val != 0 {
		t.Errorf("service_issue gauge = %v, want 0", val)
	}
	if val := testutil.ToFloat64(serviceStatusGauge.WithLabelValues("genesys-cloud", "outage")); val != 0 {
		t.Errorf("outage gauge = %v, want 0", val)
	}
}
