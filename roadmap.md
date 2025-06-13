# RSS Exporter Roadmap

This document outlines the planned enhancements to transform RSS Exporter into a robust cloud platform status scraper. Each item is tracked with a checkbox so progress can be monitored.

## Completed
- [x] Basic RSS/Atom polling and Prometheus metrics
- [x] Configuration file for defining feeds and scrape intervals
- [x] Metric tests for AWS, GCP, Genesys, and OpenAI feeds

## Planned

### Provider-Specific Parsing
- [x] Implement dedicated parsers for AWS, GCP, Azure, and other providers to handle unique feed formats
- [x] Extend `parseAWSGUID` to include additional AWS feed structures
- [x] Provide an interface so each provider reports service name, region, and status consistently

### Feed Deduplication
- [x] Enhance `incidentKey` or create provider-specific deduplication logic for repeated incident entries

### Error Handling & Retries
- [x] Add retry logic with exponential backoff around feed fetching
- [x] Surface errors through Prometheus metrics and logs

### Concurrency & Rate Control
- [ ] Use worker pools or contexts to manage goroutines and enable graceful shutdown

### Configuration Improvements
- [ ] Allow specifying provider type per service in `config.yml` to select the appropriate parser

### Documentation & Examples
- [ ] Expand README with instructions for new provider modules and configuration samples
- [ ] Add troubleshooting section for common issues

