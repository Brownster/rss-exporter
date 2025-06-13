package main

import "fmt"

// defaultAWSServiceFeeds returns a list of AWS region RSS feeds.
//
// AWS consolidated its status RSS feeds and provides a single
// "multipleservices" feed per region. This function returns all
// currently available regional feeds.
func defaultAWSServiceFeeds() []ServiceFeed {
	regions := []string{
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
		"ca-central-1",
		"ca-west-1",
		"sa-east-1",
		"eu-central-1",
		"eu-central-2",
		"eu-north-1",
		"eu-south-1",
		"eu-south-2",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"ap-east-1",
		"ap-east-2",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-northeast-3",
		"ap-south-1",
		"ap-south-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-southeast-3",
		"ap-southeast-4",
		"ap-southeast-5",
		"ap-southeast-7",
		"me-central-1",
		"me-south-1",
		"af-south-1",
		"il-central-1",
		"mx-central-1",
		"us-gov-east-1",
		"us-gov-west-1",
		"cn-north-1",
		"cn-northwest-1",
	}

	feeds := make([]ServiceFeed, 0, len(regions))
	for _, region := range regions {
		name := fmt.Sprintf("aws_%s", region)
		url := fmt.Sprintf("https://status.aws.amazon.com/rss/multipleservices-%s.rss", region)
		feeds = append(feeds, ServiceFeed{Name: name, URL: url, Interval: 300})
	}
	return feeds
}
