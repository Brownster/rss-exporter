package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const (
	namespace                     = "rss_exporter"
	defaultTimeout                = 10 * time.Second
	defaultConnectTimeout         = 5 * time.Second
	defaultAllowedMaxRedirection  = 0
	defaultValidStatus            = "200"
	defaultLogRespBodyLengthLimit = 2048
)

var (
	probePhases        = []string{"dns", "connect", "tls", "first_resp_byte", "transfer", "all"}
	serviceStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_status",
			Help:      "Current service status parsed from configured feeds.",
		},
		[]string{"service", "state"},
	)
)

type ProbeRSSOpts struct {
	// Target specifies the URL to probe.
	// It corresponds to the 'target' HTTP parameter.
	Target string

	// Timeout defines the overall duration limit for the probe,
	// setting the maximum time it can run.
	// It corresponds to the 'timeout' HTTP parameter.
	Timeout time.Duration

	// MaxRedirection is the maximum number of redirects to follow.
	// TODO: Define and implement redirection handling.
	MaxRedirection int

	// LogRespBodyOnFailure controls whether the response body is logged if the probe fails.
	// Response body content will be truncated if reached maximum length and encoded in base64.
	// Default maximum length of body is 2048 characters.
	// It corresponds to the 'log_resp_body_on_failure' HTTP parameter.
	LogRespBodyOnFailure bool

	// FailOnEmptyItems specifies whether an empty set of feed items should trigger a failure condition.
	// It corresponds to the 'fail_on_empty_items' HTTP parameter.
	FailOnEmptyItems bool

	// ValidStatusCodes is a comma-separated string of HTTP status codes considered valid.
	// If a probe response's status code is not in this list, the probe is marked as failed.
	// It corresponds to the 'valid_status' HTTP parameter.
	ValidStatusCodes []int

	// PeekRespHeadersLog is a comma-separated string of response header names.
	// For each specified header, its first value will be peeked from the response and logged.
	// It corresponds to the 'peek_resp_headers_log' HTTP parameter.
	PeekRespHeadersLog []string

	// PeekResponseHeadersLabels defines response headers to peek and write to labels.
	// TODO: Implement this feature to peek specified response headers and write them as labels.
	PeekRespHeadersLabels []string

	Registry *prometheus.Registry
}

func newProbeRSSOpts() (opts *ProbeRSSOpts) {
	return &ProbeRSSOpts{
		ValidStatusCodes:      make([]int, 0),
		PeekRespHeadersLog:    make([]string, 0),
		PeekRespHeadersLabels: make([]string, 0),
	}
}

func probeRSS(opts *ProbeRSSOpts, logger *logrus.Entry) {

	logger.Tracef("new probe request: %+v", opts)
	var (
		rssProbeDurationSec = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "probe_duration_seconds",
				Help:      "Duration of http request by phase.",
			},
			[]string{"target", "phase"},
		)
		rssProbeStatusCode = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "probe_http_status_code",
				Help:      "HTTP status code of the RSS feed probe.",
			},
			[]string{"target"},
		)
		rssItemCount = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "items_count",
				Help:      "Number of items in the fetched RSS feed.",
			},
			[]string{"target"},
		)
		rssProbeSuccess = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "probe_success",
				Help:      "Indicates if the RSS feed probe was successful (1) or failed (0).",
			},
			[]string{"target"},
		)
		rssLatestItemPubTsSec = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "latest_item_published_timestamp_seconds",
				Help:      "Timestamp of the most recently published item in the feed, in Unix seconds.",
			},
			[]string{"target"},
		)
		rssFeedContentSizeBytes = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "feed_content_size_bytes",
				Help:      "Size of the feed content, in bytes.",
			},
			[]string{"target"},
		)
	)
	opts.Registry.MustRegister(rssProbeDurationSec)
	opts.Registry.MustRegister(rssProbeStatusCode)
	opts.Registry.MustRegister(rssItemCount)
	opts.Registry.MustRegister(rssProbeSuccess)
	opts.Registry.MustRegister(rssLatestItemPubTsSec)
	opts.Registry.MustRegister(rssFeedContentSizeBytes)
	logger.Traceln("prometheus metrics registered")

	// Initialization to ensure values of all phases is 0
	for _, phase := range probePhases {
		rssProbeDurationSec.WithLabelValues(opts.Target, phase).Set(0)
	}
	rssProbeSuccess.WithLabelValues(opts.Target).Set(0)
	rssProbeStatusCode.WithLabelValues(opts.Target).Set(0)
	rssItemCount.WithLabelValues(opts.Target).Set(0)
	rssLatestItemPubTsSec.WithLabelValues(opts.Target).Set(0)
	logger.Traceln("prometheus metrics initialized")

	// Default Transport
	defaultTransport := http.DefaultTransport.(*http.Transport).Clone()
	defaultTransport.DialContext = (&net.Dialer{Timeout: defaultConnectTimeout}).DialContext

	// Create RTT Tracer Transport
	customTransport := newRSSTransport(defaultTransport)

	client := &http.Client{
		Transport: customTransport,
		Timeout:   opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= opts.MaxRedirection {
				logger.Traceln("reached max redirection times, using last response")
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", opts.Target, nil)
	if err != nil {
		logger.Errorf("error creating request: %v", err)
		return
	}
	req.Header.Set("User-Agent", "RSS-Exporter/1.0.0")

	logger.Traceln("making probe request")
	resp, err := client.Do(req)

	if err == nil {
		// Set received response headers time before read body
		customTransport.ReceivedResponseHeaders()
	}

	rssProbeDurationSec.WithLabelValues(opts.Target, "dns").Set(customTransport.current.DNSDuration())
	rssProbeDurationSec.WithLabelValues(opts.Target, "connect").Set(customTransport.current.ConnectDuration())
	rssProbeDurationSec.WithLabelValues(opts.Target, "tls").Set(customTransport.current.TLSDuration())
	logger.Debugf("dns duration: %.2f, connect duration: %.2f, tls duration: %.2f",
		customTransport.current.DNSDuration(),
		customTransport.current.ConnectDuration(),
		customTransport.current.TLSDuration())

	if err != nil {
		logger.Errorf("error during HTTP Do: %v", err)
		customTransport.End()
		rssProbeDurationSec.WithLabelValues(opts.Target, "all").Set(customTransport.current.AllDuration())
		return
	}
	rssProbeStatusCode.WithLabelValues(opts.Target).Set(float64(resp.StatusCode))
	defer resp.Body.Close()

	customTransport.RespBodyReadStart()
	body, err := io.ReadAll(resp.Body)
	customTransport.RespBodyReadDone()
	customTransport.End()
	logger = logger.WithField("http_code", resp.StatusCode)

	// Peek header values to log
	if len(opts.PeekRespHeadersLog) > 0 {
		logFields := logrus.Fields{}
		for _, logHdr := range opts.PeekRespHeadersLog {
			logFields[strings.ReplaceAll(logHdr, "-", "_")] = resp.Header.Get(logHdr)
		}
		logger = logger.WithFields(logFields)
	}

	rssProbeDurationSec.WithLabelValues(opts.Target, "first_resp_byte").Set(customTransport.current.FirstResponseByteDuration())
	rssProbeDurationSec.WithLabelValues(opts.Target, "transfer").Set(customTransport.current.RespBodyTransferDuration())
	rssProbeDurationSec.WithLabelValues(opts.Target, "all").Set(customTransport.current.AllDuration())
	rssFeedContentSizeBytes.WithLabelValues(opts.Target).Set(float64(len(body)))
	logger.Debugf("http duration: %.2f", customTransport.current.AllDuration())

	if !slices.Contains(opts.ValidStatusCodes, resp.StatusCode) {
		logger = wrapFailureBody(logger, body)
		logger.Warnf("got invalid status code '%d' while probing", resp.StatusCode)
		return
	}

	if err != nil {
		logger = wrapFailureBody(logger, body)
		logger.Errorf("error during read response body: %v", err)
		return
	}

	logger.Traceln("parsing feed content")
	fp := gofeed.NewParser()
	feed, err := fp.Parse(bytes.NewReader(body))

	if err != nil {
		logger = wrapFailureBody(logger, body)
		logger.Errorf("an error occurred while parsing body: %v", err)
		return
	}

	logger.Debugf("parsed %d feed items", len(feed.Items))
	rssItemCount.WithLabelValues(opts.Target).Set(float64(len(feed.Items)))

	if opts.FailOnEmptyItems && len(feed.Items) == 0 {
		logger = wrapFailureBody(logger, body)
		rssProbeSuccess.WithLabelValues(opts.Target).Set(0)
	} else {
		rssProbeSuccess.WithLabelValues(opts.Target).Set(1)
	}

	logger.Traceln("parsing latest item time")
	var latestItemTime *time.Time
	for _, item := range feed.Items {
		var currentItemTime *time.Time
		if item.UpdatedParsed != nil {
			currentItemTime = item.UpdatedParsed
		} else if item.PublishedParsed != nil {
			currentItemTime = item.PublishedParsed
		}

		if currentItemTime != nil {
			if latestItemTime == nil || currentItemTime.After(*latestItemTime) {
				latestItemTime = currentItemTime
			}
		}
	}

	if latestItemTime != nil {
		logger.Debugf("latest item time: %d", latestItemTime.Unix())
		rssLatestItemPubTsSec.WithLabelValues(opts.Target).Set(float64(latestItemTime.Unix()))
	} else {
		logger.Warnln("could not determine latest item published timestamp (no items or items lack dates).")
	}

	logger.Infoln("probe completed")
}

func probeHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "missing param 'target'", http.StatusBadRequest)
		return
	}

	logger := logrus.WithFields(logrus.Fields{
		"target": target,
	})

	urlQueries := r.URL.Query()

	// Timeout
	probeTimeout := time.Duration(appConfig.DefaultTimeout) * time.Second
	timeoutVal, err := parseQuery(&urlQueries, "timeout", appConfig.DefaultTimeout)
	if err != nil {
		logger.Warnf("Invalid timeout value: %v. Using default: %s", err, probeTimeout)
	} else {
		if timeoutVal <= 0 {
			logger.Warnf("Invalid timeout value '%d' (must be positive). Using default: %s", timeoutVal, probeTimeout)
		} else {
			probeTimeout = time.Duration(timeoutVal) * time.Second
		}
	}

	// Valid Status
	validStatusCodes, _ := parseStringSlice(defaultValidStatus, strconv.Atoi)
	validStatusStr := urlQueries.Get("valid_status")
	if validStatusStr != "" {
		validStatus, err := parseStringSlice(validStatusStr, strconv.Atoi)
		if err != nil {
			logger.Warnf("Error parsing 'valid_status': %v. Using default.", err)
		} else {
			validStatusCodes = validStatus
		}
	}

	// TODO: Max Redirection
	maxRedirection := defaultAllowedMaxRedirection

	// Peek response headers log
	var respHdrs = []string{}
	respHdrsLogStr := urlQueries.Get("peek_resp_headers_log")
	if respHdrsLogStr != "" {
		respHdrs, _ = parseStringSlice(respHdrsLogStr, func(s string) (string, error) { return s, nil })
	}

	registry := prometheus.NewRegistry()
	opts := newProbeRSSOpts()
	opts.Target = target
	opts.Timeout = probeTimeout
	opts.Registry = registry
	opts.MaxRedirection = maxRedirection
	opts.ValidStatusCodes = validStatusCodes
	opts.PeekRespHeadersLog = respHdrs
	opts.LogRespBodyOnFailure = len(urlQueries.Get("log_resp_body_on_failure")) > 0
	opts.FailOnEmptyItems = len(urlQueries.Get("fail_on_empty_items")) > 0
	probeRSS(opts, logger)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func wrapFailureBody(logger *logrus.Entry, body []byte) *logrus.Entry {
	if len(body) > defaultLogRespBodyLengthLimit {
		body = body[:defaultLogRespBodyLengthLimit]
	}

	return logger.WithField("resp_body", base64.StdEncoding.EncodeToString(body))
}

func parseStringSlice[T comparable](strs string, parseFunc func(string) (T, error)) (ret []T, err error) {
	ret = make([]T, 0)
	for str := range strings.SplitSeq(strs, ",") {
		if len(str) == 0 {
			continue
		}
		var sc T
		sc, err = parseFunc(str)
		if err != nil {
			err = fmt.Errorf("invalid item '%s' to parse: %v", str, err)
			ret = make([]T, 0)
			return
		}
		ret = append(ret, sc)
	}
	return
}

func parseQuery(q *url.Values, key string, defVal int) (v int, err error) {
	v = defVal
	str := q.Get(key)
	if str != "" {
		var val int
		val, err = strconv.Atoi(str)
		if err != nil {
			return
		}
		v = val
	}
	return
}

func extractServiceStatus(item *gofeed.Item) (service string, state string, active bool) {
	text := strings.ToUpper(strings.TrimSpace(item.Title + " " + item.Description))
	switch {
	case strings.Contains(text, "RESOLVED"):
		state = "resolved"
	case strings.Contains(text, "OUTAGE"):
		state = "outage"
	case strings.Contains(text, "SERVICE ISSUE"), strings.Contains(text, "SERVICE IMPACT"):
		state = "service_issue"
	}

	if state == "" {
		return
	}

	service = strings.TrimSpace(item.Title)
	active = state != "resolved"
	return
}
