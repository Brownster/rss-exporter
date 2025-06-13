# Metrics

The exporter exposes the following Prometheus metrics:

| Metric | Labels | Description |
|--------|--------|-------------|
| `rss_exporter_service_status` | `service`, `customer` (optional), `state` | Current service state: `ok`, `service_issue`, or `outage`. |
| `rss_exporter_service_issue_info` | `service`, `customer` (optional), `service_name`, `region`, `title`, `link`, `guid` | Information about the active incident, value is always `1` when present. |
| `rss_exporter_fetch_errors_total` | `service`, `customer` (optional) | Counter of consecutive feed fetch failures. |

Example scrape output:

```text
# HELP rss_exporter_service_status Current service status parsed from configured feeds.
# TYPE rss_exporter_service_status gauge
rss_exporter_service_status{service="openai",state="ok"} 1
rss_exporter_service_status{service="openai",state="outage"} 0
```

