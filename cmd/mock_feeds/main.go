package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/aws.rss", awsHandler)
	http.HandleFunc("/azure.rss", azureHandler)
	http.HandleFunc("/gcp.atom", gcpHandler)
	http.HandleFunc("/cloudflare.atom", cloudflareHandler)
	http.HandleFunc("/genesys.atom", genesysHandler)
	http.HandleFunc("/openai.atom", openaiHandler)

	addr := ":8000"
	log.Printf("mock feed server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func randChoice(list []string) string {
	return list[rand.Intn(len(list))]
}

// AWS feed generator (RSS)
func awsHandler(w http.ResponseWriter, r *http.Request) {
	services := []string{"ec2", "s3", "rds", "lambda"}
	regions := []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"}
	state := randChoice([]string{"issue", "outage", "resolved"})
	svc := randChoice(services)
	region := randChoice(regions)

	titleState := strings.ToUpper(state)
	if state == "resolved" {
		titleState = "RESOLVED"
	} else if state == "issue" {
		titleState = "SERVICE ISSUE"
	} else {
		titleState = "OUTAGE"
	}

	ts := time.Now().Format(time.RFC1123Z)
	guid := fmt.Sprintf("%s-%s_%s", svc, region, state)
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title><![CDATA[Amazon %s (%s) Service Status]]></title>
    <link>https://status.aws.amazon.com/</link>
    <language>en-us</language>
    <lastBuildDate>%s</lastBuildDate>
    <generator>Mock RSS Generator</generator>
    <description><![CDATA[Mock feed]]></description>
    <ttl>5</ttl>
    <item>
      <title><![CDATA[%s: %s %s]]></title>
      <link>https://status.aws.amazon.com/</link>
      <pubDate>%s</pubDate>
      <guid isPermaLink="false">https://status.aws.amazon.com/#%s</guid>
      <description><![CDATA[Simulated %s for %s in %s]]></description>
    </item>
  </channel>
</rss>`, strings.Title(svc), region, ts, titleState, strings.Title(svc), region, ts, guid, state, svc, region)

	w.Header().Set("Content-Type", "application/rss+xml")
	fmt.Fprint(w, content)
}

// Azure feed generator (RSS)
func azureHandler(w http.ResponseWriter, r *http.Request) {
	services := []string{"storage", "compute", "sql", "appservice"}
	regions := []string{"eastus", "westus2", "northeurope", "southeastasia"}
	state := randChoice([]string{"issue", "outage", "resolved"})
	svc := randChoice(services)
	region := randChoice(regions)

	status := "Service issue"
	if state == "outage" {
		status = "Service outage"
	} else if state == "resolved" {
		status = "Service issue" // resolved item includes "resolved" in GUID only
	}

	ts := time.Now().Format(time.RFC1123Z)
	guid := fmt.Sprintf("%s-%s_%s", svc, region, state)
	itemState := status
	if state == "resolved" {
		itemState = "Service issue"
	}
	content := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<rss version="2.0">
  <channel>
    <title>Azure Status</title>
    <item>
      <title>%s: %s - %s</title>
      <link>https://status.azure.com/en-us/status</link>
      <pubDate>%s</pubDate>
      <guid>%s</guid>
      <description>Simulated %s for %s in %s</description>
    </item>
  </channel>
</rss>`, itemState, strings.Title(svc), strings.Title(region), ts, guid, state, svc, region)

	w.Header().Set("Content-Type", "application/rss+xml")
	fmt.Fprint(w, content)
}

// GCP feed generator (Atom)
func gcpHandler(w http.ResponseWriter, r *http.Request) {
	services := []string{"Compute Engine", "Cloud Storage", "BigQuery"}
	state := randChoice([]string{"issue", "outage", "resolved"})
	svc := randChoice(services)

	status := "SERVICE ISSUE"
	if state == "outage" {
		status = "SERVICE OUTAGE"
	} else if state == "resolved" {
		status = "RESOLVED"
	}

	ts := time.Now().Format(time.RFC3339)
	id := fmt.Sprintf("mock-%d", rand.Intn(100000))

	content := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Google Cloud Service Health Updates</title>
  <updated>%s</updated>
  <link href="https://status.cloud.google.com/" rel="alternate" type="text/html"/>
  <link href="https://status.cloud.google.com/en/feed.atom" rel="self"/>
  <author><name>Google Cloud</name></author>
  <id>https://status.cloud.google.com/</id>
  <entry>
    <title>%s: %s</title>
    <link href="https://status.cloud.google.com/incidents/%s" rel="alternate" type="text/html"/>
    <id>tag:status.cloud.google.com,2025:feed:%s</id>
    <updated>%s</updated>
    <summary type="html"><p>Simulated %s incident.</p></summary>
  </entry>
</feed>`, ts, status, svc, id, id, ts, state)

	w.Header().Set("Content-Type", "application/atom+xml")
	fmt.Fprint(w, content)
}

// Cloudflare feed generator (Atom)
func cloudflareHandler(w http.ResponseWriter, r *http.Request) {
	services := []string{"Cloudflare CDN", "DNS", "Workers"}
	state := randChoice([]string{"issue", "outage", "resolved"})
	svc := randChoice(services)
	status := "SERVICE ISSUE"
	if state == "outage" {
		status = "SERVICE OUTAGE"
	} else if state == "resolved" {
		status = "RESOLVED"
	}
	ts := time.Now().Format(time.RFC3339)
	id := fmt.Sprintf("mock-%d", rand.Intn(100000))

	content := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Cloudflare Status</title>
  <updated>%s</updated>
  <entry>
    <title>%s: %s</title>
    <id>https://www.cloudflarestatus.com/incidents/%s</id>
    <link href="https://www.cloudflarestatus.com/incidents/%s"/>
    <updated>%s</updated>
    <content>%s incident for %s</content>
  </entry>
</feed>`, ts, status, svc, id, id, ts, state, svc)

	w.Header().Set("Content-Type", "application/atom+xml")
	fmt.Fprint(w, content)
}

// Genesys feed generator (Atom)
func genesysHandler(w http.ResponseWriter, r *http.Request) {
	services := []string{"Text to Speech", "Contact Center", "Voice"}
	state := randChoice([]string{"issue", "outage", "resolved"})
	svc := randChoice(services)
	status := "Service issue"
	if state == "outage" {
		status = "Outage"
	} else if state == "resolved" {
		status = "Resolved"
	}
	ts := time.Now().Format(time.RFC3339)
	id := fmt.Sprintf("mock-%d", rand.Intn(100000))

	content := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<feed xml:lang="en-US" xmlns="http://www.w3.org/2005/Atom">
  <title>Genesys Cloud Status - Incident History</title>
  <updated>%s</updated>
  <entry>
    <title>%s: %s</title>
    <id>tag:status.mypurecloud.com,2005:%s</id>
    <updated>%s</updated>
    <content type="html"><p>%s incident for %s</p></content>
  </entry>
</feed>`, ts, status, svc, id, ts, state, svc)

	w.Header().Set("Content-Type", "application/atom+xml")
	fmt.Fprint(w, content)
}

// OpenAI feed generator (Atom)
func openaiHandler(w http.ResponseWriter, r *http.Request) {
	services := []string{"API", "ChatGPT", "Embeddings"}
	state := randChoice([]string{"issue", "outage", "resolved"})
	svc := randChoice(services)
	status := "Service issue"
	if state == "outage" {
		status = "Outage"
	} else if state == "resolved" {
		status = "Resolved"
	}
	ts := time.Now().Format(time.RFC3339)
	id := fmt.Sprintf("mock-%d", rand.Intn(100000))

	content := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>OpenAI status</title>
  <updated>%s</updated>
  <entry>
    <title>%s: %s</title>
    <id>https://status.openai.com/incidents/%s</id>
    <link href="https://status.openai.com/incidents/%s"/>
    <updated>%s</updated>
    <summary type="html">%s incident for %s</summary>
  </entry>
</feed>`, ts, status, svc, id, id, ts, state, svc)

	w.Header().Set("Content-Type", "application/atom+xml")
	fmt.Fprint(w, content)
}
