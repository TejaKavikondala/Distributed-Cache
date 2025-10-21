package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	cache "github.com/TejaKavikondala/Distributed-Cache/pkg/cache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "Cache server address")
	numGoroutines := flag.Int("goroutines", 100, "Number of concurrent goroutines")
	requestsPerGoroutine := flag.Int("requests", 1000, "Requests per goroutine")
	flag.Parse()

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	client := cache.NewCacheServiceClient(conn)

	totalRequests := *numGoroutines * *requestsPerGoroutine
	var successCount int64
	var failureCount int64

	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < *numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < *requestsPerGoroutine; j++ {
				key := fmt.Sprintf("key:%d:%d", id, j)
				value := []byte(fmt.Sprintf("value_%d_%d", id, j))

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

				_, err := client.Set(ctx, &cache.SetRequest{
					Key:        key,
					Value:      value,
					TtlSeconds: 3600,
				})

				if err != nil {
					atomic.AddInt64(&failureCount, 1)
					cancel()
					continue
				}

				resp, err := client.Get(ctx, &cache.GetRequest{Key: key})
				if err != nil || !resp.Found {
					atomic.AddInt64(&failureCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}

				cancel()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Println("\n=== Load Test Results ===")
	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failureCount)
	fmt.Printf("Duration: %v\n", duration)
	if successCount > 0 {
		fmt.Printf("Throughput: %.2f requests/sec\n", float64(successCount)/duration.Seconds())
		fmt.Printf("Avg Latency: %.2f ms\n", (duration.Seconds()*1000)/float64(successCount))
	}
}
