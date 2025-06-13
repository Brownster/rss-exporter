# RSS Exporter

RSS Exporter periodically polls configured RSS or Atom feeds and exposes Prometheus metrics.

## Usage

Build and run:

```bash
go build -o rss_exporter .
./rss_exporter -config /path/to/config.yml
```

Metrics are available at `http://<listen_address>:<listen_port>/metrics`.

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
  - name: aws
    url: https://status.aws.amazon.com/rss/all.rss
    interval: 300
```

The `services` section lists feeds to poll. `interval` is in seconds.

## Exposed Metrics

* `rss_exporter_service_status{service="<name>",state="<status>"}` - Current state of each service (`ok`, `service_issue`, `outage`).
* `rss_exporter_service_issue_info{service="<name>",title="<item_title>",link="<item_link>",guid="<item_guid>"}` - Set to `1` while a service reports an active issue.

## Example output:

# HELP rss_exporter_service_issue_info Details for the currently active service issue.
# TYPE rss_exporter_service_issue_info gauge
rss_exporter_service_issue_info{guid="https://status.openai.com//incidents/01JTS3ZKAK8KDEER57AZ0AEE6T",link="https://status.openai.com//incidents/01JTS3ZKAK8KDEER57AZ0AEE6T",service="openai",title="WhatsApp 1-800-CHATGPT partial outage"} 1
# HELP rss_exporter_service_status Current service status parsed from configured feeds.
# TYPE rss_exporter_service_status gauge
rss_exporter_service_status{service="aws_apigateway_eu-central-1",state="ok"} 1
rss_exporter_service_status{service="aws_apigateway_eu-central-1",state="outage"} 0
rss_exporter_service_status{service="aws_apigateway_eu-central-1",state="service_issue"} 0
rss_exporter_service_status{service="aws_athena_us-west-2",state="ok"} 1
rss_exporter_service_status{service="aws_athena_us-west-2",state="outage"} 0
rss_exporter_service_status{service="aws_athena_us-west-2",state="service_issue"} 0
rss_exporter_service_status{service="aws_connect_eu-west-2",state="ok"} 1
rss_exporter_service_status{service="aws_connect_eu-west-2",state="outage"} 0
rss_exporter_service_status{service="aws_connect_eu-west-2",state="service_issue"} 0
rss_exporter_service_status{service="azure",state="ok"} 1
rss_exporter_service_status{service="azure",state="outage"} 0
rss_exporter_service_status{service="azure",state="service_issue"} 0
rss_exporter_service_status{service="cloudflare",state="ok"} 1
rss_exporter_service_status{service="cloudflare",state="outage"} 0
rss_exporter_service_status{service="cloudflare",state="service_issue"} 0
rss_exporter_service_status{service="gcp",state="ok"} 1
rss_exporter_service_status{service="gcp",state="outage"} 0
rss_exporter_service_status{service="gcp",state="service_issue"} 0
rss_exporter_service_status{service="genesys-cloud",state="ok"} 1
rss_exporter_service_status{service="genesys-cloud",state="outage"} 0
rss_exporter_service_status{service="genesys-cloud",state="service_issue"} 0
rss_exporter_service_status{service="okta",state="ok"} 1
rss_exporter_service_status{service="okta",state="outage"} 0
rss_exporter_service_status{service="okta",state="service_issue"} 0
rss_exporter_service_status{service="openai",state="ok"} 0
rss_exporter_service_status{service="openai",state="outage"} 1
rss_exporter_service_status{service="openai",state="service_issue"} 0
