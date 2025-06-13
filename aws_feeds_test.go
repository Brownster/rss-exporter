package main

import "testing"

func TestDefaultAWSServiceFeeds(t *testing.T) {
	feeds := defaultAWSServiceFeeds()
	if len(feeds) == 0 {
		t.Fatal("no feeds returned")
	}
	want := "https://status.aws.amazon.com/rss/apigateway-eu-central-1.rss"
	found := false
	for _, f := range feeds {
		if f.URL == want {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected feed %s not found", want)
	}
}
