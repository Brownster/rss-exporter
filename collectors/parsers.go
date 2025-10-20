package collectors

import (
	"strings"

	"github.com/mmcdole/gofeed"
)

// Scraper extracts provider-specific information from a feed item and
// also provides a deduplication key used to filter repeated entries.
type Scraper interface {
	// ServiceInfo returns the service name and region associated with the item.
	ServiceInfo(item *gofeed.Item) (serviceName, region string)
	// IncidentKey returns a stable identifier for the incident represented
	// by this item. Items with the same key will be deduplicated.
	IncidentKey(item *gofeed.Item) string
}

type awsParser struct{}

func (awsParser) ServiceInfo(item *gofeed.Item) (string, string) {
	return ParseAWSGUID(item.GUID)
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

type twilioParser struct{}

func (twilioParser) ServiceInfo(item *gofeed.Item) (string, string) {
	title := strings.TrimSpace(item.Title)
	if title == "" {
		return "", ""
	}

	trimSuffixes := []string{
		" Maintenance",
		" maintenance",
		" Incident",
		" incident",
		" Update",
		" update",
	}
	cleaned := title
	for _, suffix := range trimSuffixes {
		cleaned = strings.TrimSpace(strings.TrimSuffix(cleaned, suffix))
	}

	separators := []string{" Twilio ", " SMS ", " MMS ", " Voice ", " Account "}
	for _, sep := range separators {
		if idx := strings.Index(cleaned, sep); idx != -1 {
			region := strings.TrimSpace(cleaned[:idx])
			service := strings.TrimSpace(cleaned[idx:])
			service = strings.TrimPrefix(service, "Twilio ")
			service = strings.TrimSpace(service)
			return service, region
		}
	}

	return cleaned, ""
}

func (twilioParser) IncidentKey(item *gofeed.Item) string {
	return genericParser{}.IncidentKey(item)
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

// ScraperForService selects a scraper based on the provider or service name.
func ScraperForService(provider, service string) Scraper {
	p := strings.ToLower(provider)
	switch p {
	case "aws":
		return awsParser{}
	case "gcp":
		return gcpParser{}
	case "azure":
		return azureParser{}
	case "twilio":
		return twilioParser{}
	case "":
		// fall back to service name when provider not set
	default:
		if p != "" {
			return genericParser{}
		}
	}

	svc := strings.ToLower(service)
	switch {
	case strings.Contains(svc, "aws"):
		return awsParser{}
	case strings.Contains(svc, "gcp"):
		return gcpParser{}
	case strings.Contains(svc, "azure"):
		return azureParser{}
	case strings.Contains(svc, "twilio"):
		return twilioParser{}
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

// ParseAWSGUID extracts the AWS service name and region from a GUID string.
// GUIDs may appear in several formats, including:
//
//	https://status.aws.amazon.com/#service-region_12345
//	arn:aws:health:region::event/AWS_SERVICE_eventid
//
// Unknown formats return empty strings.
func ParseAWSGUID(guid string) (serviceName, region string) {
	if idx := strings.Index(guid, "#"); idx != -1 {
		guid = guid[idx+1:]
	}

	if strings.HasPrefix(guid, "arn:aws:health:") {
		// arn:aws:health:region::event/AWS_SERVICENAME_foo
		parts := strings.Split(guid, ":")
		if len(parts) >= 4 {
			region = parts[3]
		}
		if idx := strings.LastIndex(guid, "/"); idx != -1 {
			svc := guid[idx+1:]
			svc = strings.TrimPrefix(svc, "AWS_")
			svcParts := strings.SplitN(svc, "_", 2)
			serviceName = strings.ToLower(svcParts[0])
		}
		return
	}

	if idx := strings.IndexAny(guid, "_"); idx != -1 {
		guid = guid[:idx]
	}

	parts := strings.Split(guid, "-")
	if len(parts) < 2 {
		return "", ""
	}

	if len(parts) >= 3 {
		region = strings.Join(parts[len(parts)-3:], "-")
		serviceName = strings.Join(parts[:len(parts)-3], "-")
	} else {
		region = parts[len(parts)-1]
		serviceName = strings.Join(parts[:len(parts)-1], "-")
	}
	serviceName = strings.ToLower(serviceName)
	return
}
