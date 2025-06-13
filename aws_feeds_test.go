package main

import "testing"

func TestDefaultAWSServiceFeeds(t *testing.T) {
	feeds := defaultAWSServiceFeeds()
	if len(feeds) != 2 {
		t.Fatalf("expected 2 feeds got %d", len(feeds))
	}

	want := map[string]string{
		"aws_apigateway_eu-central-1": "https://status.aws.amazon.com/rss/apigateway-eu-central-1.rss",
		"aws_connect_eu-west-2":       "https://status.aws.amazon.com/rss/connect-eu-west-2.rss",
	}

	for _, f := range feeds {
		url, ok := want[f.Name]
		if !ok {
			t.Errorf("unexpected feed name %s", f.Name)
			continue
		}
		if f.URL != url {
			t.Errorf("feed %s url got %s want %s", f.Name, f.URL, url)
		}
		delete(want, f.Name)
	}
	if len(want) != 0 {
		t.Errorf("feeds missing: %v", want)
	}
}
