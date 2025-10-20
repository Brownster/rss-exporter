package collectors

import (
	"regexp"
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

type genesysParser struct{}

func (genesysParser) ServiceInfo(item *gofeed.Item) (string, string) {
	service, region := parseGenesysTitle(strings.TrimSpace(item.Title))
	if service == "" {
		service = inferGenesysService(item)
	}
	region = determineGenesysRegion(region, item)

	return slugify(service), region
}

func (genesysParser) IncidentKey(item *gofeed.Item) string {
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
	case "genesyscloud":
		return genesysParser{}
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
	case strings.Contains(svc, "genesys"):
		return genesysParser{}
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

var (
	genesysRegionKeyReplacer = strings.NewReplacer(
		"(", " ",
		")", " ",
		",", " ",
		".", " ",
		":", " ",
		"/", " ",
		"-", " ",
		"_", " ",
		"*", " ",
	)
	slugifyCleanupRegexp      = regexp.MustCompile(`[^a-z0-9]+`)
	genesysTitleRegionPattern = regexp.MustCompile(`(?i)([A-Za-z0-9&/ ]+\([^\)]+\))`)
)

type genesysRegionInfo struct {
	label string
	group string
}

var genesysRegionMap = map[string]genesysRegionInfo{
	"americas":           {label: "americas", group: "americas"},
	"americas us east":   {label: "americas-us-east", group: "americas"},
	"us east":            {label: "americas-us-east", group: "americas"},
	"americas us west":   {label: "americas-us-west", group: "americas"},
	"us west":            {label: "americas-us-west", group: "americas"},
	"canada":             {label: "americas-canada", group: "americas"},
	"americas canada":    {label: "americas-canada", group: "americas"},
	"sao paulo":          {label: "americas-sao-paulo", group: "americas"},
	"americas sao paulo": {label: "americas-sao-paulo", group: "americas"},
	"virginia":           {label: "americas-virginia", group: "americas"},
	"ohio":               {label: "americas-ohio", group: "americas"},
	"oregon":             {label: "americas-oregon", group: "americas"},
	"montreal":           {label: "americas-montreal", group: "americas"},
	"fedramp":            {label: "americas-fedramp", group: "americas"},
	"emea":               {label: "emea", group: "emea"},
	"dublin":             {label: "emea-dublin", group: "emea"},
	"frankfurt":          {label: "emea-frankfurt", group: "emea"},
	"london":             {label: "emea-london", group: "emea"},
	"ireland":            {label: "emea-ireland", group: "emea"},
	"paris":              {label: "emea-paris", group: "emea"},
	"zurich":             {label: "emea-zurich", group: "emea"},
	"emea dublin":        {label: "emea-dublin", group: "emea"},
	"emea frankfurt":     {label: "emea-frankfurt", group: "emea"},
	"emea london":        {label: "emea-london", group: "emea"},
	"emea paris":         {label: "emea-paris", group: "emea"},
	"emea zurich":        {label: "emea-zurich", group: "emea"},
	"apac":               {label: "apac", group: "apac"},
	"tokyo":              {label: "apac-tokyo", group: "apac"},
	"osaka":              {label: "apac-osaka", group: "apac"},
	"sydney":             {label: "apac-sydney", group: "apac"},
	"hong kong":          {label: "apac-hong-kong", group: "apac"},
	"singapore":          {label: "apac-singapore", group: "apac"},
	"mumbai":             {label: "apac-mumbai", group: "apac"},
	"jakarta":            {label: "apac-jakarta", group: "apac"},
	"seoul":              {label: "apac-seoul", group: "apac"},
	"apac tokyo":         {label: "apac-tokyo", group: "apac"},
	"apac sydney":        {label: "apac-sydney", group: "apac"},
	"apac mumbai":        {label: "apac-mumbai", group: "apac"},
	"apac osaka":         {label: "apac-osaka", group: "apac"},
	"cape town":          {label: "afs-cape-town", group: "afs"},
	"afs":                {label: "afs", group: "afs"},
	"mea":                {label: "mea", group: "mea"},
	"uae":                {label: "mea-uae", group: "mea"},
	"dubai":              {label: "mea-uae", group: "mea"},
	"global":             {label: "global", group: "global"},
	"multiple regions":   {label: "multiple-regions", group: "multiple"},
	"cross regional":     {label: "multiple-regions", group: "multiple"},
	"cross region":       {label: "multiple-regions", group: "multiple"},
}

var genesysServiceKeywords = []struct {
	keyword string
	label   string
}{
	{keyword: "acd email", label: "Email"},
	{keyword: "email routing", label: "Email"},
	{keyword: "email", label: "Email"},
	{keyword: "sms", label: "SMS"},
	{keyword: "mms", label: "SMS"},
	{keyword: "whatsapp", label: "WhatsApp"},
	{keyword: "web messaging", label: "Web Messaging"},
	{keyword: "messaging", label: "Messaging"},
	{keyword: "voice", label: "Voice"},
	{keyword: "contact center", label: "Contact Center"},
	{keyword: "social listening", label: "Social Listening"},
	{keyword: "recording", label: "Recording"},
	{keyword: "recordings", label: "Recording"},
	{keyword: "analytics", label: "Analytics"},
	{keyword: "reporting", label: "Analytics"},
	{keyword: "text to speech", label: "Speech"},
	{keyword: "speech to text", label: "Speech"},
	{keyword: "outbound dialing", label: "Outbound Dialing"},
	{keyword: "workforce management", label: "Workforce Management"},
}

func parseGenesysTitle(title string) (service, region string) {
	service = strings.TrimSpace(title)
	if service == "" {
		return "", ""
	}

	service = strings.Join(strings.Fields(service), " ")

	if matches := genesysTitleRegionPattern.FindAllStringIndex(service, -1); len(matches) > 0 {
		last := matches[len(matches)-1]
		region = strings.TrimSpace(service[last[0]:last[1]])
		service = strings.TrimSpace(service[:last[0]] + service[last[1]:])
	}

	if idx := strings.LastIndex(service, " - "); idx != -1 {
		if region == "" {
			region = strings.TrimSpace(service[idx+3:])
		}
		service = strings.TrimSpace(service[:idx])
	}

	if idx := strings.Index(service, ":"); idx != -1 {
		service = strings.TrimSpace(service[idx+1:])
	}

	service = trimGenesysServiceSuffix(service)
	return service, strings.TrimSpace(region)
}

func trimGenesysServiceSuffix(service string) string {
	if service == "" {
		return service
	}
	lower := strings.ToLower(service)
	suffixes := []string{
		" incident",
		" incidents",
		" service issue",
		" service issues",
		" service impact",
		" services impacted",
		" services impaired",
		" update",
		" outage",
		" partial outage",
		" monitoring",
	}
	for {
		trimmed := false
		for _, suffix := range suffixes {
			if strings.HasSuffix(lower, suffix) {
				service = strings.TrimSpace(service[:len(service)-len(suffix)])
				lower = strings.ToLower(service)
				trimmed = true
			}
		}
		if !trimmed {
			break
		}
	}
	return strings.TrimSpace(service)
}

func inferGenesysService(item *gofeed.Item) string {
	text := strings.ToLower(strings.Join([]string{item.Title, item.Description, item.Content}, " "))
	if text == "" {
		return ""
	}

	seen := make(map[string]struct{})
	matches := []string{}
	for _, candidate := range genesysServiceKeywords {
		if strings.Contains(text, candidate.keyword) {
			if _, ok := seen[candidate.label]; ok {
				continue
			}
			seen[candidate.label] = struct{}{}
			matches = append(matches, candidate.label)
		}
	}
	if len(matches) == 0 {
		return ""
	}
	if len(matches) == 1 {
		return matches[0]
	}
	return "Multiple Services"
}

func determineGenesysRegion(initial string, item *gofeed.Item) string {
	fallback := strings.TrimSpace(initial)
	regions := map[genesysRegionInfo]struct{}{}

	if info, ok := lookupGenesysRegion(initial); ok {
		regions[info] = struct{}{}
	}

	text := strings.Join([]string{item.Title, item.Description, item.Content}, " ")
	for info := range detectGenesysRegions(text) {
		regions[info] = struct{}{}
	}

	if len(regions) == 0 {
		lower := strings.ToLower(text)
		switch {
		case strings.Contains(lower, "global"):
			return "global"
		case strings.Contains(lower, "cross regional"), strings.Contains(lower, "multiple regions"), strings.Contains(lower, "cross-region"):
			return "multiple-regions"
		}
		return slugify(fallback)
	}

	groups := make(map[string]struct{})
	var first genesysRegionInfo
	count := 0
	for info := range regions {
		groups[info.group] = struct{}{}
		if count == 0 {
			first = info
		}
		count++
	}

	if len(groups) > 1 {
		return "multiple-regions"
	}

	if count == 1 {
		return first.label
	}

	for group := range groups {
		if group == "multiple" {
			return "multiple-regions"
		}
		if group != "" {
			return group
		}
	}

	if first.label != "" {
		return first.label
	}
	return slugify(fallback)
}

func lookupGenesysRegion(region string) (genesysRegionInfo, bool) {
	key := normalizeGenesysRegionKey(region)
	info, ok := genesysRegionMap[key]
	return info, ok
}

func detectGenesysRegions(text string) map[genesysRegionInfo]struct{} {
	cleaned := " " + normalizeGenesysRegionKey(text) + " "
	found := make(map[genesysRegionInfo]struct{})
	for key, info := range genesysRegionMap {
		if strings.Contains(cleaned, " "+key+" ") {
			found[info] = struct{}{}
		}
	}
	return found
}

func normalizeGenesysRegionKey(value string) string {
	if value == "" {
		return ""
	}
	cleaned := genesysRegionKeyReplacer.Replace(strings.ToLower(value))
	return strings.Join(strings.Fields(cleaned), " ")
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	value = slugifyCleanupRegexp.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	return value
}
