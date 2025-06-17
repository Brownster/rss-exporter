package collectors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMetricsEndpoint(t *testing.T) {
	metricsMu.Lock()
	metricsData = map[string]*serviceMetrics{
		"test": {State: "ok"},
	}
	metricsMu.Unlock()
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	mux.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("unexpected status code %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "rss_exporter_service_status") {
		t.Errorf("service status metric missing")
	}
}
