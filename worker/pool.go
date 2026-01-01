package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/hrittikhere/mongodb-benchmark/db"

	"go.mongodb.org/mongo-driver/mongo"
)

func StartPool(ctx context.Context, client *mongo.Client, workerCount int) {
	var wg sync.WaitGroup

	log.Printf("Starting worker pool with %d workers (Max Throughput)", workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			reqID := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
					qCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					db.ExecuteQuery(qCtx, client, reqID)
					cancel()
					reqID++
				}
			}
		}(i)
	}

	wg.Wait()
}
