package main

import (
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"
)

// roundTripTrace holds timings for a single HTTP roundtrip.
// refer to net/http/httptrace docs for more details
type roundTripTrace struct {
	tls              bool
	start            time.Time // will be set when a new DNS Dial begins or a new connection's Dial begins.
	end              time.Time // will be set after body read
	dnsStart         time.Time // will be set when a new DNS Dial begins.
	dnsDone          time.Time // will be set when a new DNS Dial completes.
	connectStart     time.Time // will be set when a new connection's Dial begins.
	connectDone      time.Time // will be set called when a new connection's Dial completes.
	tlsStart         time.Time // will be set when tls handshake begins.
	tlsDone          time.Time // will be set when tls handshake completes.
	reqHeadersDone   time.Time // will be set after written all request headers.
	respHeadersStart time.Time // will be set when the first byte of the response headers is available.
	respHeadersDone  time.Time // will be set after received all response headers.
	respBodyStart    time.Time // will be set when body read begins.
	respBodyDone     time.Time // will be set after body read.
}

type rssTransport struct {
	Transport http.RoundTripper
	mu        sync.Mutex
	current   *roundTripTrace
}

func newRSSTransport(rt http.RoundTripper) *rssTransport {
	return &rssTransport{
		Transport: rt,
	}
}

func (rtt *roundTripTrace) DNSDuration() (d float64) {
	if !rtt.dnsStart.IsZero() && !rtt.dnsDone.IsZero() {
		d = rtt.dnsDone.Sub(rtt.dnsStart).Seconds()
	}
	return
}

func (rtt *roundTripTrace) ConnectDuration() (d float64) {
	if !rtt.connectStart.IsZero() && !rtt.connectDone.IsZero() {
		d = rtt.connectDone.Sub(rtt.connectStart).Seconds()
	}
	return
}

func (rtt *roundTripTrace) TLSDuration() (d float64) {
	if rtt.tls && !rtt.tlsStart.IsZero() && !rtt.tlsDone.IsZero() {
		d = rtt.tlsDone.Sub(rtt.tlsStart).Seconds()
	}
	return
}

func (rtt *roundTripTrace) FirstResponseByteDuration() (d float64) {
	if !rtt.start.IsZero() && !rtt.respHeadersStart.IsZero() {
		return rtt.respHeadersStart.Sub(rtt.start).Seconds()
	}
	return
}

func (rtt *roundTripTrace) RespBodyTransferDuration() (d float64) {
	if !rtt.respHeadersDone.IsZero() && !rtt.end.IsZero() {
		d = rtt.end.Sub(rtt.respHeadersDone).Seconds()
	}
	return
}

func (rtt *roundTripTrace) AllDuration() (d float64) {
	if !rtt.start.IsZero() && !rtt.end.IsZero() {
		d = rtt.end.Sub(rtt.start).Seconds()
	}
	return
}

// RoundTrip switches to a new trace, then runs embedded RoundTripper.
func (t *rssTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	trace := &roundTripTrace{}
	if req.URL.Scheme == "https" {
		trace.tls = true
	}

	t.mu.Lock()
	t.current = trace
	t.mu.Unlock()

	// Hook
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
		DNSStart:             t.DNSStart,
		DNSDone:              t.DNSDone,
		ConnectStart:         t.ConnectStart,
		ConnectDone:          t.ConnectDone,
		TLSHandshakeStart:    t.TLSHandshakeStart,
		TLSHandshakeDone:     t.TLSHandshakeDone,
		WroteHeaders:         t.WroteRequestHeaders,
		GotFirstResponseByte: t.GotFirstResponseByte,
	}))

	return t.Transport.RoundTrip(req)
}

func (t *rssTransport) DNSStart(_ httptrace.DNSStartInfo) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.dnsStart = time.Now()

		// t.current.start should be set here or set in ConnectStart
		if t.current.start.IsZero() {
			t.current.start = t.current.dnsStart
		}
	}
}

func (t *rssTransport) DNSDone(_ httptrace.DNSDoneInfo) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.dnsDone = time.Now()
	}
}

func (t *rssTransport) ConnectStart(_, _ string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.connectStart = time.Now()

		// DNS phase will be skipped if connect to IP
		if t.current.dnsDone.IsZero() {
			t.current.dnsDone = t.current.connectStart
		}

		// Due to this func may be called multiple times,
		// so we only need to set it on first time,
		// refer to net/http/httptrace for more details.
		if t.current.start.IsZero() {
			t.current.start = t.current.connectStart
		}
	}
}

func (t *rssTransport) ConnectDone(net, addr string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.connectDone = time.Now()
	}
}

func (t *rssTransport) GotFirstResponseByte() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.respHeadersStart = time.Now()
	}
}

func (t *rssTransport) ReceivedResponseHeaders() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.respHeadersDone = time.Now()
	}
}

func (t *rssTransport) TLSHandshakeStart() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.tlsStart = time.Now()
	}
}

func (t *rssTransport) TLSHandshakeDone(_ tls.ConnectionState, _ error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.tlsDone = time.Now()
	}
}

func (t *rssTransport) WroteRequestHeaders() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.reqHeadersDone = time.Now()
	}
}

func (t *rssTransport) RespBodyReadStart() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.respBodyStart = time.Now()
	}
}

func (t *rssTransport) RespBodyReadDone() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.respBodyDone = time.Now()
	}
}

func (t *rssTransport) End() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.end = time.Now()
	}
}
