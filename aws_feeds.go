package main

import "fmt"

// defaultAWSServiceFeeds returns a list of AWS service RSS feeds.
func defaultAWSServiceFeeds() []ServiceFeed {
	services := []string{
		"apigateway",
		"ec2",
		"s3",
		"rds",
		"lambda",
		"dynamodb",
	}
	regions := []string{
		"us-east-1",
		"us-west-2",
		"eu-central-1",
		"eu-west-1",
	}

	feeds := make([]ServiceFeed, 0, len(services)*len(regions))
	for _, svc := range services {
		for _, region := range regions {
			name := fmt.Sprintf("aws_%s_%s", svc, region)
			url := fmt.Sprintf("https://status.aws.amazon.com/rss/%s-%s.rss", svc, region)
			feeds = append(feeds, ServiceFeed{Name: name, URL: url, Interval: 300})
		}
	}
	return feeds
}
