package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/4O4-Not-F0und/rss-exporter/internal/exporter"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	wg := exporter.StartWorkers(ctx, exporter.AppConfig.Services)

	mux := http.NewServeMux()
	mux.HandleFunc("/", landingPageHandler)
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", exporter.AppConfig.ListenAddress, exporter.AppConfig.ListenPort),
		Handler: mux,
	}

	go func() {
		logrus.Infof("Listening at: %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("http server failed: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("server shutdown: %v", err)
	}

	wg.Wait()
}

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
