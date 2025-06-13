package exporter

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func loadAzureIssueFeed(t *testing.T) []byte {
	t.Helper()
	data, err := ioutil.ReadFile("testdata/azure_issue.rss")
	if err != nil {
		t.Fatalf("read feed: %v", err)
	}
	return data
}

func TestUpdateServiceStatus_AzureIssue(t *testing.T) {
	data := loadAzureIssueFeed(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer ts.Close()

	metricsMu.Lock()
	metricsData = map[string]*serviceMetrics{}
	metricsMu.Unlock()

	cfg := ServiceFeed{Name: "azure-storage", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	metricsMu.Lock()
	sm := metricsData["azure-storage"]
	metricsMu.Unlock()

	if sm.State != "service_issue" {
		t.Errorf("state = %v, want service_issue", sm.State)
	}
	if sm.Issue == nil || sm.Issue.ServiceName != "storage" || sm.Issue.Region != "eastus" {
		t.Errorf("issue info mismatch: %+v", sm.Issue)
	}
}
