package collectors

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

func loadOpenAIResolvedFeed(t *testing.T) *gofeed.Feed {
	t.Helper()
	data, err := os.ReadFile("testdata/openai_resolved.atom")
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
	data, err := os.ReadFile("testdata/openai_resolved.atom")
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

	cfg := ServiceFeed{Name: "openai", URL: ts.URL, Interval: 0}
	updateServiceStatus(cfg, logrus.NewEntry(logrus.New()))

	metricsMu.Lock()
	sm := metricsData["openai"]
	metricsMu.Unlock()

	if sm.State != "ok" {
		t.Errorf("state = %v, want ok", sm.State)
	}
}
