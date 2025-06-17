package collectors

import (
	"context"
	"sync"
)

// StartWorkers launches a monitor goroutine for each service and
// returns a WaitGroup that waits for all workers to exit.
func StartWorkers(ctx context.Context, services []ServiceFeed) *sync.WaitGroup {
	var wg sync.WaitGroup
	for _, svc := range services {
		wg.Add(1)
		go func(s ServiceFeed) {
			defer wg.Done()
			monitorService(ctx, s)
		}(svc)
	}
	return &wg
}
