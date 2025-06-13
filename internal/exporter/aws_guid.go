package exporter

import "strings"

// parseAWSGUID extracts the AWS service name and region from a GUID string.
// GUIDs may appear in several formats, including:
//
//	https://status.aws.amazon.com/#service-region_12345
//	arn:aws:health:region::event/AWS_SERVICE_eventid
//
// Unknown formats return empty strings.
func parseAWSGUID(guid string) (serviceName, region string) {
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
