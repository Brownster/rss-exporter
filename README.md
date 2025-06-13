# RSS Exporter

RSS Exporter periodically polls configured RSS or Atom feeds and exposes Prometheus metrics.

## Usage

Build and run:

```bash
go build -o rss_exporter .
./rss_exporter -config /path/to/config.yml -aws
```

Metrics are available at `http://<listen_address>:<listen_port>/metrics`.

The optional `-aws` flag enables a predefined set of AWS status feeds.

## Configuration

Example `config.yml`:

```yaml
listen_address: 127.0.0.1
listen_port: 9091
log_level: info
services:
  - name: gcp
    url: https://status.cloud.google.com/en/feed.atom
    interval: 300
```

The `services` section lists feeds to poll. `interval` is in seconds.

## Exposed Metrics

* `rss_exporter_service_status{service="<name>",state="<status>"}` - Current state of each service (`ok`, `service_issue`, `outage`).
* `rss_exporter_service_issue_info{service="<name>",title="<item_title>",link="<item_link>",guid="<item_guid>"}` - Set to `1` while a service reports an active issue.

