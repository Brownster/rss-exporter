package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestProbeMaxRedirections(t *testing.T) {
	var reqCount int
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount++
		w.Header().Set("Location", ts.URL)
		w.WriteHeader(http.StatusFound)
	}))
	defer ts.Close()

	req := httptest.NewRequest("GET", fmt.Sprintf("/probe?target=%s&max_redirections=1", url.QueryEscape(ts.URL)), nil)
	rr := httptest.NewRecorder()

	probeHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", rr.Code)
	}
	if reqCount != 2 {
		t.Errorf("expected 2 requests, got %d", reqCount)
	}

	body := rr.Body.String()
	expectedMetric := fmt.Sprintf("rss_exporter_probe_http_status_code{target=\"%s\"} 302", ts.URL)
	if !strings.Contains(body, expectedMetric) {
		t.Errorf("status metric missing or incorrect: %s", body)
	}
}
