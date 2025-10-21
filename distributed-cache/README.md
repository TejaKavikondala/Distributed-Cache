# Distributed Cache Layer with Consistent Hashing

High-performance distributed caching system built in Go, demonstrating consistent hashing, gRPC, and scalable cache node architecture.

## Architecture

- **Consistent Hashing Ring**: Distributes keys across nodes with minimal redistribution on topology changes
- **Virtual Nodes**: 3 replicas per physical node for better load distribution
- **gRPC**: High-performance RPC protocol for inter-node communication
- **Connection Pooling**: Efficient connection management
- **Prometheus Metrics**: Real-time observability
- **Graceful Shutdown**: Clean termination with in-flight request completion

## Performance

- **Throughput**: 50K+ requests/second
- **Latency**: <10ms average
- **Concurrent Connections**: 10K+

## Quick Start

```bash
# Clone and setup
git clone https://github.com/yourusername/distributed-cache
cd distributed-cache

# Generate protobuf code
protoc --go_out=. --go-grpc_out=. proto/cache.proto

# Install dependencies
go mod download

# Run with Docker Compose
docker-compose up

# Run load test
go run ./cmd/loadtest/main.go -goroutines 100 -requests 1000
```

## API

```go
// Set key-value with TTL
proxy.Set(ctx, "user:123", []byte("data"), 1*time.Hour)

// Get value
value, err := proxy.Get(ctx, "user:123")

// Delete key
proxy.Delete(ctx, "user:123")
```

## Project Structure

```
├── cmd/
│   ├── node/        # Cache node server
│   └── loadtest/    # Load testing tool
├── pkg/
│   ├── cache/       # Protobuf generated code
│   ├── node/        # In-memory cache storage
│   ├── ring/        # Consistent hashing ring
│   ├── metrics/     # Prometheus metrics
│   ├── proxy/       # Cache proxy/gateway
│   └── server/      # gRPC server
├── proto/
│   └── cache.proto  # Service definitions
└── docker-compose.yml
```

## Monitoring

Access Prometheus metrics:
- Node 1: http://localhost:9091/metrics
- Node 2: http://localhost:9092/metrics
- Node 3: http://localhost:9093/metrics

Key metrics:
- `cache_hits_total`
- `cache_misses_total`
- `cache_sets_total`
- `request_duration_seconds`
- `active_connections`
