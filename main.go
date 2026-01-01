package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/hrittikhere/mongodb-benchmark/db"
	"github.com/hrittikhere/mongodb-benchmark/metrics"
	"github.com/hrittikhere/mongodb-benchmark/worker"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	str := getEnv(key, "")
	if str == "" {
		return fallback
	}
	if v, err := strconv.Atoi(str); err == nil {
		return v
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	str := getEnv(key, "")
	if str == "" {
		return fallback
	}
	if v, err := strconv.ParseFloat(str, 64); err == nil {
		return v
	}
	return fallback
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it. Using environment variables/defaults.")
	}

	dbURI := getEnv("DB_CONNECTION_STRING", "")
	if dbURI == "" {
		log.Fatal("DB_CONNECTION_STRING is not set. Please set it in .env or as an environment variable.")
	}

	workerCount := getEnvInt("WORKER_COUNT", 10)
	dbMaxPool := getEnvInt("DB_MAX_POOL_SIZE", 20)

	if dbMaxPool < workerCount {
		log.Printf("[WARN] DB_MAX_POOL_SIZE (%d) is less than WORKER_COUNT (%d). Increasing pool to %d", dbMaxPool, workerCount, workerCount)
		dbMaxPool = workerCount
	}

	metrics.Init(prometheus.DefaultRegisterer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Connecting to MongoDB...")
	setupCtx, setupCancel := context.WithTimeout(ctx, 10*time.Second)
	defer setupCancel()

	client, err := db.Connect(setupCtx, dbURI, uint64(dbMaxPool))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting DB: %v", err)
		}
	}()
	log.Println("Connected to MongoDB successfully.")

	go func() {
		port := getEnv("PORT", "8080")
		addr := ":" + port
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Metrics server listening on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		worker.StartPool(ctx, client, workerCount)
	}()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Printf("Workers: %d | Goroutines: %d", workerCount, runtime.NumGoroutine())
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down...")
	cancel()

	time.Sleep(2 * time.Second)
	log.Println("Shutdown complete.")
}
