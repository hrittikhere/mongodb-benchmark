package db

import (
	"context"
	"log"
	"time"

	"github.com/hrittikhere/mongodb-benchmark/metrics"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(ctx context.Context, uri string, maxPoolSize uint64) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(uri)
	opts.SetMaxPoolSize(maxPoolSize)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

func ExecuteQuery(ctx context.Context, client *mongo.Client, requestID int) {
	collection := client.Database("bench").Collection("data")

	start := time.Now()
	baseText := "lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur excepteur sint occaecat cupidatat non proident sunt in culpa qui officia deserunt mollit anim id est laborum"

	doc := bson.D{
		{"req_id", requestID},
		{"timestamp", time.Now()},
		{"payload", baseText},
		{"payload_2", baseText},
	}
	_, err := collection.InsertOne(ctx, doc)
	metrics.DbQueryLatencySeconds.WithLabelValues("insert").Observe(time.Since(start).Seconds())
	metrics.DbQueryTotal.WithLabelValues("insert").Inc()

	if err != nil {
		metrics.DbErrorsTotal.WithLabelValues("insert").Inc()
		log.Printf("Insert error [req=%d]: %v", requestID, err)
		return
	}

	start = time.Now()
	var result bson.M
	err = collection.FindOne(ctx, bson.D{{"req_id", requestID}}).Decode(&result)
	metrics.DbQueryLatencySeconds.WithLabelValues("find").Observe(time.Since(start).Seconds())
	metrics.DbQueryTotal.WithLabelValues("find").Inc()

	if err != nil {
		metrics.DbErrorsTotal.WithLabelValues("find").Inc()
		log.Printf("Find error [req=%d]: %v", requestID, err)
	}

	if requestID%10 == 0 {
		log.Printf("Request #%d: Insert+Find completed", requestID)
	}
}
