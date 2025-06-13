# Architecture Overview

This document describes the internal structure of **RSS Exporter** and how the main components interact.

## Package layout

```
rss-exporter/
├── main.go             # Application entry point
├── internal/exporter/   # Core exporter logic
│   ├── config.go        # Configuration loader
│   ├── fetch.go         # Feed fetching with retries
│   ├── provider.go      # Provider-specific parsing
│   ├── service.go       # Feed monitoring worker
│   ├── metrics.go       # Prometheus metrics collector
│   └── worker.go        # Goroutine orchestration
```

All production code lives in the `internal/exporter` package. Tests and sample feed files are kept alongside the implementation.

## Main flow

1. **Configuration** is loaded from YAML using `initConfig` in `config.go`.
2. `main.go` creates a context and starts a worker goroutine for each configured feed via `StartWorkers`.
3. Each worker continuously fetches its feed in `monitorService`. Feed retrieval is retried using exponential backoff (`fetchFeedWithRetry`).
4. Feed items are parsed by a provider-specific parser chosen by `parserForService`. Parsed status information is stored in shared metrics structures.
5. Prometheus metrics are exposed via the `/metrics` HTTP handler.

## Adding new providers

Implement the `ItemParser` interface with `ServiceInfo` and `IncidentKey`. Update `parserForService` to return the new parser when the provider name is requested. Unit tests under `internal/exporter` demonstrate expected behaviour for existing providers.

