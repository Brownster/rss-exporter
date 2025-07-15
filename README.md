# RSS Exporter

RSS Exporter periodically polls configured RSS or Atom feeds and exposes Prometheus metrics. It is built on a minimal version of the company's `monitoring-maas` framework which schedules each feed scraper automatically.

## Usage

Build and run:

```bash
go build -o rss_exporter .
./rss_exporter -config.file=/path/to/config.yml
```

Metrics are available at `http://<listen_address>:<listen_port>/metrics`.

## Configuration

Example `config.yml`:

```yaml
listen_address: 127.0.0.1
listen_port: 9091
log_level: info
services:
  - name: aws
    provider: aws
    # customer 
    url: https://status.aws.amazon.com/rss/all.rss
    interval: 300
  - name: aws
    provider: aws
    customer: some_customer 
    url: https://status.aws.amazon.com/rss/some_customer-specific.rss
    interval: 300
  - name: gcp
    provider: gcp
    # customer defaults to the service name
    url: https://status.cloud.google.com/en/feed.atom
    interval: 300
  - name: genesys-cloud
    provider: genesyscloud
    url: https://status.mypurecloud.com/history.atom
    interval: 300
  - name: azure
    provider: azure
    url: https://azurestatuscdn.azureedge.net/en-gb/status/feed
    interval: 300
  - name: cloudflare
    provider: cloudflare
    url: https://www.cloudflarestatus.com/history.atom
    interval: 300
  - name: openai
    provider: openai
    url: https://status.openai.com/history.atom
    interval: 300
```

The `services` section lists feeds to poll. `interval` is in seconds. Each entry
can optionally specify a `provider` to explicitly select the scraper used for
that service. When omitted, the provider is inferred from the service name.

### Provider Modules

The exporter includes dedicated scrapers for several cloud providers:

* **aws** – parses AWS Health RSS feeds and extracts service and region.
* **gcp** – handles Google Cloud status feeds.
* **azure** – parses Azure status feeds and extracts service and region.

Any other value falls back to the generic scraper. Provider names like
`cloudflare`, `genesyscloud`, `okta`, or `openai` all use the generic collector.
When the `provider` field is omitted, the service name is inspected to select a
suitable scraper.

## Exposed Metrics

* `rss_exporter_service_status{service="<name>",customer="<customer>",state="<status>"}` - Current state of each service (`ok`, `service_issue`, `outage`).
* `rss_exporter_service_issue_info{service="<name>",customer="<customer>",service_name="<service>",region="<region>",title="<item_title>",link="<item_link>",guid="<item_guid>"}` - Set to `1` while a service reports an active issue. The `service_name` and `region` labels are populated only for providers that supply them (e.g. AWS and Azure).

## Example output:

# HELP rss_exporter_service_issue_info Details for the currently active service issue.
# TYPE rss_exporter_service_issue_info gauge
rss_exporter_service_issue_info{guid="storage-eastus_issue",link="https://status.azure.com/en-us/status",region="eastus",service="azure",service_name="storage",title="Service issue: Storage - East US"} 1
# HELP rss_exporter_service_status Current service status parsed from configured feeds.
# TYPE rss_exporter_service_status gauge
rss_exporter_service_status{service="azure",state="ok"} 0
rss_exporter_service_status{service="azure",state="outage"} 0
rss_exporter_service_status{service="azure",state="service_issue"} 1

### Local Testing With Sample Feeds

The repository includes sample RSS and Atom files under `testdata/` that can be served locally for quick testing. Use a temporary configuration like:

```yaml
listen_address: 127.0.0.1
listen_port: 9095
log_level: debug
services:
  - name: openai
    provider: openai
    url: http://localhost:8000/openai_resolved.atom
    interval: 1
  - name: azure
    provider: azure
    url: http://localhost:8000/azure_issue.rss
    interval: 1
```

Running the exporter with these feeds yields metrics such as:

```text
# HELP rss_exporter_service_issue_info Details for the currently active service issue.
# TYPE rss_exporter_service_issue_info gauge
rss_exporter_service_issue_info{guid="storage-eastus_issue",link="https://status.azure.com/en-us/status",region="eastus",service="azure",service_name="storage",title="Service issue: Storage - East US"} 1
# HELP rss_exporter_service_status Current service status parsed from configured feeds.
# TYPE rss_exporter_service_status gauge
rss_exporter_service_status{service="azure",state="ok"} 0
rss_exporter_service_status{service="azure",state="outage"} 0
rss_exporter_service_status{service="azure",state="service_issue"} 1
rss_exporter_service_status{service="openai",state="ok"} 1
rss_exporter_service_status{service="openai",state="outage"} 0
rss_exporter_service_status{service="openai",state="service_issue"} 0
```

### Mock Feed Server

For more dynamic testing you can run the included mock feed server which
generates random incidents for several providers:

```bash
go run ./cmd/mock_feeds
```

Point the exporter at `config.mock.yml` to scrape these endpoints and iterate on dashboards or alerts.

## Graceful Shutdown

The exporter relies on the `maas` framework which gracefully stops all scheduled
scrapers when the process receives `SIGINT` or `SIGTERM`.
