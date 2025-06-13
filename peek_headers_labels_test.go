package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestProbeHeadersLabels(t *testing.T) {
	data := loadAWSFeed(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "testserver")
		w.Header().Set("ETag", "abc123")
		w.Write(data)
	}))
	defer ts.Close()

	req := httptest.NewRequest("GET", fmt.Sprintf("/probe?target=%s&peek_resp_headers_labels=Server,ETag", url.QueryEscape(ts.URL)), nil)
	rr := httptest.NewRecorder()

	probeHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "rss_exporter_probe_headers_info") {
		t.Fatalf("metric missing in response: %s", body)
	}

	expected := fmt.Sprintf("rss_exporter_probe_headers_info{ETag=\"abc123\",Server=\"testserver\",target=\"%s\"} 1", ts.URL)
	if !strings.Contains(body, expected) {
		t.Errorf("expected metric %q not found in response. body: %s", expected, body)
	}
}
