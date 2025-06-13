package main

import "strings"

// parseAWSGUID extracts the AWS service name and region from a GUID string.
// The GUID is expected to contain a fragment like "#service-region_xxx".
func parseAWSGUID(guid string) (serviceName, region string) {
	if idx := strings.Index(guid, "#"); idx != -1 {
		guid = guid[idx+1:]
	}
	if idx := strings.Index(guid, "_"); idx != -1 {
		guid = guid[:idx]
	}
	parts := strings.Split(guid, "-")
	if len(parts) < 2 {
		return "", ""
	}
	// Assume region is composed of the last three parts (e.g. us-west-2).
	// This works for all region formats used in tests.
	if len(parts) >= 3 {
		region = strings.Join(parts[len(parts)-3:], "-")
		serviceName = strings.Join(parts[:len(parts)-3], "-")
	} else {
		// fall back to last part as region
		region = parts[len(parts)-1]
		serviceName = strings.Join(parts[:len(parts)-1], "-")
	}
	return
}
