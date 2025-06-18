package fetcher

import (
	"context"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

const (
	defaultTimeout      = 10 * time.Second
	defaultFetchRetries = 3
)

// FetchFeedWithRetry retrieves the feed URL with exponential backoff retries.
func FetchFeedWithRetry(url string, logger *logrus.Entry) (*gofeed.Feed, error) {
	backoff := time.Second
	var lastErr error
	for i := 1; i <= defaultFetchRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		feed, err := gofeed.NewParser().ParseURLWithContext(url, ctx)
		cancel()
		if err == nil {
			return feed, nil
		}
		lastErr = err
		logger.Debugf("attempt %d failed: %v", i, err)
		if i < defaultFetchRetries {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	return nil, lastErr
}
