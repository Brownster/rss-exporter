# Architecture Overview

This document describes the internal structure of **RSS Exporter** and how the main components interact.

## Package layout

```
rss-exporter/
├── cmd/rss_exporter/   # Application entry point
│   └── main.go
├── collectors/         # Exporter logic and scrapers
│   ├── config.go       # Configuration loader
│   ├── feed.go         # Feed monitoring worker
│   ├── parsers.go      # Scraper implementations
│   ├── exporter.go     # Wrapper around workers and HTTP server
│   └── testdata/       # Sample feed files
├── connectors/         # Maas compatible connectors
│   ├── http.go         # HTTP connector implementing maas.Connector
│   └── http_mock.go    # Test helper for mocks
└── internal/connectors/ # Feed fetching helpers
    └── connector.go    # HTTP fetch with retries
```

All production code lives in the `collectors` package. Tests and sample feed files are kept alongside the implementation.

## Main flow

1. **Configuration** is loaded from YAML using `initConfig` in `config.go`.
2. `main.go` creates a context and starts a worker goroutine for each configured feed via `StartWorkers`.
3. Each worker continuously fetches its feed in `monitorService`. Feed retrieval is retried using exponential backoff via the connector.
4. Feed items are parsed by a provider-specific scraper chosen by `ScraperForService`. Parsed status information is stored in shared metrics structures.
5. Prometheus metrics are exposed via the `/metrics` HTTP handler.

## Adding new providers

Implement the `Scraper` interface with `ServiceInfo` and `IncidentKey`. Update `ScraperForService` to return the new scraper when the provider name is requested. Unit tests under `collectors` demonstrate expected behaviour for existing providers.

