package main

import (
	"strings"

	"github.com/mmcdole/gofeed"
)

// ItemParser extracts provider-specific information from a feed item.
type ItemParser interface {
	// ServiceInfo returns the service name and region associated with the item.
	ServiceInfo(item *gofeed.Item) (serviceName, region string)
}

type awsParser struct{}

func (awsParser) ServiceInfo(item *gofeed.Item) (string, string) {
	return parseAWSGUID(item.GUID)
}

type gcpParser struct{}

func (gcpParser) ServiceInfo(item *gofeed.Item) (string, string) {
	// GCP feeds don't expose a service or region in a structured way.
	return "", ""
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

type genericParser struct{}

func (genericParser) ServiceInfo(item *gofeed.Item) (string, string) {
	return "", ""
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
