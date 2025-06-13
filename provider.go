package main

import (
	"strings"

	"github.com/mmcdole/gofeed"
)

// ItemParser extracts provider-specific information from a feed item.
// ItemParser extracts provider-specific information from a feed item and
// also provides a deduplication key used to filter repeated entries.
type ItemParser interface {
	// ServiceInfo returns the service name and region associated with the item.
	ServiceInfo(item *gofeed.Item) (serviceName, region string)
	// IncidentKey returns a stable identifier for the incident represented
	// by this item. Items with the same key will be deduplicated.
	IncidentKey(item *gofeed.Item) string
}

type awsParser struct{}

func (awsParser) ServiceInfo(item *gofeed.Item) (string, string) {
	return parseAWSGUID(item.GUID)
}

func (awsParser) IncidentKey(item *gofeed.Item) string {
	key := item.GUID
	if key == "" {
		key = item.Link
	}
	if idx := strings.Index(key, "#"); idx != -1 {
		key = key[idx+1:]
	}
	key = strings.TrimSuffix(key, "_resolved")
	key = strings.TrimSuffix(key, "_issue")
	return key
}

type gcpParser struct{}

func (gcpParser) ServiceInfo(item *gofeed.Item) (string, string) {
	// GCP feeds don't expose a service or region in a structured way.
	return "", ""
}

func (gcpParser) IncidentKey(item *gofeed.Item) string {
	if strings.Contains(item.Link, "status.cloud.google.com/incidents/") {
		return item.Link
	}
	if item.GUID != "" {
		return item.GUID
	}
	return item.Title
}

type azureParser struct{}

func (azureParser) ServiceInfo(item *gofeed.Item) (string, string) {
	if item.GUID != "" {
		if svc, reg := parseAzureGUID(item.GUID); svc != "" {
			return svc, reg
		}
	}
	title := strings.ToLower(item.Title)
	if idx := strings.Index(title, ":"); idx != -1 {
		title = strings.TrimSpace(title[idx+1:])
	}
	parts := strings.Split(title, " - ")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(title), ""
}

func (azureParser) IncidentKey(item *gofeed.Item) string {
	key := item.GUID
	if key == "" {
		key = item.Link
	}
	if idx := strings.Index(key, "#"); idx != -1 {
		key = key[idx+1:]
	}
	key = strings.TrimSuffix(key, "_resolved")
	key = strings.TrimSuffix(key, "_issue")
	return key
}

type genericParser struct{}

func (genericParser) ServiceInfo(item *gofeed.Item) (string, string) {
	return "", ""
}

func (genericParser) IncidentKey(item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	if item.Link != "" {
		return item.Link
	}
	return strings.TrimSpace(item.Title)
}

// parserForService selects a parser based on the configured service name.
func parserForService(service string) ItemParser {
	svc := strings.ToLower(service)
	switch {
	case strings.Contains(svc, "aws"):
		return awsParser{}
	case strings.Contains(svc, "gcp"):
		return gcpParser{}
	case strings.Contains(svc, "azure"):
		return azureParser{}
	default:
		return genericParser{}
	}
}

// parseAzureGUID extracts service name and region from an Azure GUID of the form
// "service-region_xyz". Unknown formats return empty strings.
func parseAzureGUID(guid string) (serviceName, region string) {
	if idx := strings.Index(guid, "#"); idx != -1 {
		guid = guid[idx+1:]
	}
	if idx := strings.IndexAny(guid, "_"); idx != -1 {
		guid = guid[:idx]
	}
	parts := strings.Split(guid, "-")
	if len(parts) >= 2 {
		serviceName = strings.ToLower(parts[0])
		region = strings.Join(parts[1:], "-")
	}
	return
}
