package collectors

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	maas "github.com/sabio-engineering-product/monitoring-maas"
	"github.com/sirupsen/logrus"
)

// RssExporter wraps the existing worker and HTTP server logic.
type RssExporter struct {
	srv    *http.Server
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	conn   maas.Connector
}

// NewRssExporter prepares an exporter using the provided connector.
func NewRssExporter(c maas.Connector) (*RssExporter, error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	mux := http.NewServeMux()
	mux.HandleFunc("/", landingPageHandler)
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", AppConfig.ListenAddress, AppConfig.ListenPort),
		Handler: mux,
	}
	return &RssExporter{srv: srv, ctx: ctx, cancel: cancel, conn: c}, nil
}

// Start launches worker goroutines and the HTTP server.
func (e *RssExporter) Start() {
	e.wg = StartWorkers(e.ctx, AppConfig.Services)
	go func() {
		logrus.Infof("Listening at: %s", e.srv.Addr)
		if err := e.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("http server failed: %v", err)
		}
	}()
}

// Serve blocks until shutdown and performs a graceful stop.
func (e *RssExporter) Serve() {
	<-e.ctx.Done()
	e.cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.srv.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("server shutdown: %v", err)
	}

	if e.wg != nil {
		e.wg.Wait()
	}
}

// landingPageHandler serves a small HTML page at the root path.
func landingPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	landingPageHTML := `<html>
                        <head><title>RSS Exporter</title></head>
                        <body>
                        <h1>RSS Exporter</h1>
                        <p>Metrics available at <code>/metrics</code></p>
                        </body>
                        </html>`
	w.Write([]byte(landingPageHTML))
}
