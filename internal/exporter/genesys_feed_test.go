package exporter

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

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

	metricsMu.Lock()
	metricsData = map[string]*serviceMetrics{}
	metricsMu.Unlock()

	cfg := ServiceFeed{Name: "genesys-cloud", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	metricsMu.Lock()
	sm := metricsData["genesys-cloud"]
	metricsMu.Unlock()

	if sm.State != "ok" {
		t.Errorf("state = %v, want ok", sm.State)
	}
}
