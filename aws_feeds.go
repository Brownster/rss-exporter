package main

// defaultAWSServiceFeeds returns the list of AWS RSS feeds monitored by the
// exporter. To keep unit tests simple, only a small subset of feeds is
// returned.
func defaultAWSServiceFeeds() []ServiceFeed {
	return []ServiceFeed{
		{
			Name:     "aws_apigateway_eu-central-1",
			URL:      "https://status.aws.amazon.com/rss/apigateway-eu-central-1.rss",
			Interval: 300,
		},
		{
			Name:     "aws_connect_eu-west-2",
			URL:      "https://status.aws.amazon.com/rss/connect-eu-west-2.rss",
			Interval: 300,
		},
		{
			Name:     "aws_athena_us-west-2",
			URL:      "https://status.aws.amazon.com/rss/athena-us-west-2.rss",
			Interval: 300,
		},
	}
}
