# RSS Exporter

RSS Exporter is a Prometheus exporter designed to probe RSS/Atom feeds, parse their content, and expose relevant metrics.

It's highly configurable through a YAML file and allows for dynamic probe adjustments via HTTP query parameters.

## Features

* Probe RSS and Atom feeds.
* Export key Prometheus metrics:
    * Number of feed items.
    * Timestamp of the latest item.
    * HTTP status code of the probe.
    * Overall probe success status.
    * Detailed probe duration (DNS lookup, TCP connection, TLS handshake, time to first byte, content transfer).
* Feed content size.
* Detects service incident keywords ("SERVICE ISSUE", "OUTAGE") and exposes service status metrics.
* Continuously monitors configured service feeds and updates their status metrics.
* Configuration via a YAML file for listener settings, logging, default probe behavior.
* Customizable probe behavior per request using HTTP query parameters (e.g., target URL, timeout, valid HTTP status codes).
* Logs response body (truncated and base64 encoded) on failure.
* Option to mark a probe as failed if the feed contains no items.
* Allows peeking into specified response headers for logging purposes.

## Getting Started

### Building from Source

#### Prerequisites

* Go 1.24 or later for building from source.

1.  Clone the repository:
    ```bash
    git clone https://github.com/4O4-Not-F0und/rss-exporter
    cd rss-exporter
    ```
2.  Build the exporter:
    ```bash
    go mod tidy
    go build -o rss_exporter .
    ```

#### Running the Exporter

Execute the compiled binary. By default, it looks for a `config.yml` in the current directory.
```bash
./rss_exporter
```

You can specify a different configuration file using the `-config` flag:

```bash
./rss_exporter -config /path/to/your/config.yml
```

### Docker Compose

Refer to `docker-compose.yml`.

## Configuration

The exporter is configured using a YAML file (default: `config.yml`).

### Command-line Flags

  * `-config <path>`: Path to the configuration file. Default: `config.yml`.
  * `-enable-aws-feeds`: Automatically monitor built-in AWS status RSS feeds. The exporter now tracks the consolidated `multipleservices-<region>` feeds provided by AWS.

### Configuration Options

Below are the available options for the `config.yml` file:

  * `listen_address` (default: `0.0.0.0`): The IP address on which the exporter should listen.
  * `listen_port` (default: `9191`): The port on which the exporter should listen.
  * `log_level` (default: `info`): The logging level.
      * Valid values: `"trace"`, `"debug"`, `"info"`, `"warn"`.
  * `probe_path` (default: `/probe`): The HTTP path for the probe endpoint.
  * `default_timeout` (default: `10`): The default timeout in seconds for a probe if not specified in the HTTP request.
  * `services` (optional): A list of service feeds to monitor continuously.
      * Each service entry supports:
          * `name`: A unique service name.
          * `url`: The RSS/Atom feed URL.
          * `interval`: Refresh interval in seconds (default: `300`).

For example config, please refer to `config.example.yml`.

## Usage

The exporter exposes two HTTP endpoints:

  * `/`: A simple landing page that provides an example of how to use the probe endpoint.
  * `{probe_path}` (as configured, e.g., `/probe` by default): The main endpoint for probing RSS feeds and scraping metrics.

### Probing an RSS Feed

To probe an RSS feed, send a GET request to the configured `probe_path`. The behavior of the probe can be controlled using the following HTTP query parameters:

  * `target` (string, **required**): The URL of the RSS/Atom feed to probe.
      * Example: `target=https://www.rssboard.org/files/sample-rss-2.xml`
  * `timeout` (int, optional): The timeout for this specific probe request in seconds. This overrides the `default_timeout` value set in the configuration file. If not provided, negative, or zero, the `default_timeout` from the configuration is used.
      * Example: `timeout=15`
  * `valid_status` (string, optional): A **comma-separated** list of HTTP status codes that are considered successful for the probe.
      * Default: `"200"`
      * Example: `valid_status=200,201,204`
  * `log_resp_body_on_failure` (any, optional): If this parameter is present (e.g., `log_resp_body_on_failure=1`), the response body will be logged (truncated to 2048 bytes and base64 encoded) if the probe fails (e.g., due to an invalid status code or parsing error).
      * Disabled by default (implicitly, as it requires the parameter to be present).
      * Example: `log_resp_body_on_failure=1`
  * `fail_on_empty_items` (any, optional): If this parameter is present (e.g., `fail_on_empty_items=1`), the probe will be marked as unsuccessful if the fetched RSS feed contains zero items.
      * Disabled by default (implicitly, as it requires the parameter to be present).
      * Example: `fail_on_empty_items=1`
  * `peek_resp_headers_log` (string, optional): A **comma-separated** list of response header names. The first value of each specified header will be logged. Header names are case-insensitive (standard HTTP header behavior, logged keys are modified).
      * Example: `peek_resp_headers_log=Content-Type,Last-Modified,ETag`

### Example Probe URL

Assuming the exporter is running on `localhost:9191` with the default `probe_path` (`/probe`):

```
http://localhost:9191/probe?target=https://rss.example.com/files/sample-rss-2.xml&timeout=20&valid_status=200&fail_on_empty_items=1&peek_resp_headers_log=Server,Content-Type
```

This request will:

1.  Probe the RSS feed at `https://rss.example.com/files/sample-rss-2.xml`.
2.  Set a timeout of 20 seconds for this probe.
3.  Consider only HTTP status code 200 as successful.
4.  Mark the probe as failed if the feed has no items.
5.  Log the values of `Server` and `Content-Type` response headers.

## Exposed Metrics

All metrics exposed by the exporter are prefixed with `rss_exporter_`. The `target` label in the metrics refers to the URL of the probed RSS feed.

  * `rss_exporter_probe_duration_seconds{target="<feed_url>", phase="<phase>"}` (Gauge):
    Shows the duration in seconds for different phases of the HTTP request. The `phase` label can be one of the following:

      * `dns`: DNS lookup duration.
      * `connect`: TCP connection establishment duration.
      * `tls`: TLS handshake duration (if HTTPS).
      * `first_resp_byte`: Time from the start of the request until the first byte of the response header is received.
      * `transfer`: Duration of response body transfer (from after headers received to body read complete).
      * `all`: Total duration of the probe request from start to end of body read.

  * `rss_exporter_probe_http_status_code{target="<feed_url>"}` (Gauge):
    The HTTP status code returned by the server when probing the RSS feed.

  * `rss_exporter_items_count{target="<feed_url>"}` (Gauge):
    The number of items found in the parsed RSS feed.

  * `rss_exporter_probe_success{target="<feed_url>"}` (Gauge):
    Indicates whether the RSS feed probe was successful (1) or failed (0). A probe can fail due to network errors, unexpected HTTP status codes, feed parsing errors, or if `fail_on_empty_items` is set and no items are found.

  * `rss_exporter_latest_item_published_timestamp_seconds{target="<feed_url>"}` (Gauge):
    The Unix timestamp (in seconds) of the most recently published or updated item in the feed. This metric is not set if the feed has no items or if none of the items have a parsable published or updated date.

  * `rss_exporter_feed_content_size_bytes{target="<feed_url>"}` (Gauge):
    The size of the fetched RSS feed content in bytes.
  * `rss_exporter_service_status{service="<name>",state="<status>"}` (Gauge):
    Tracks the current state of each configured service feed. `state` can be
    `ok`, `service_issue`, or `outage`.
