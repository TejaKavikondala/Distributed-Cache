package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	cache "github.com/TejaKavikondala/Distributed-Cache/pkg/cache"
	"github.com/TejaKavikondala/Distributed-Cache/pkg/metrics"
	"github.com/TejaKavikondala/Distributed-Cache/pkg/node"
	"google.golang.org/grpc"
)

type CacheServer struct {
	cache   *node.Node
	metrics *metrics.Metrics
	port    string
	server  *grpc.Server
	cache.UnimplementedCacheServiceServer
}

func New(port string, m *metrics.Metrics) *CacheServer {
	return &CacheServer{
		cache:   node.New(),
		metrics: m,
		port:    port,
	}
}

func (s *CacheServer) Set(ctx context.Context, req *cache.SetRequest) (*cache.SetResponse, error) {
	start := time.Now()
	defer func() {
		s.metrics.RequestDuration.Observe(time.Since(start).Seconds())
	}()

	if req.Key == "" {
		return &cache.SetResponse{Success: false, Message: "key cannot be empty"}, nil
	}

	tl := time.Duration(req.TtlSeconds) * time.Second
	if tl == 0 {
		tl = 24 * time.Hour
	}

	s.cache.Set(req.Key, req.Value, tl)
	s.metrics.CacheSets.Inc()

	return &cache.SetResponse{Success: true, Message: "stored"}, nil
}

func (s *CacheServer) Get(ctx context.Context, req *cache.GetRequest) (*cache.GetResponse, error) {
	start := time.Now()
	defer func() {
		s.metrics.RequestDuration.Observe(time.Since(start).Seconds())
	}()

	value, ok := s.cache.Get(req.Key)
	if !ok {
		s.metrics.CacheMisses.Inc()
		return &cache.GetResponse{Found: false}, nil
	}

	s.metrics.CacheHits.Inc()
	return &cache.GetResponse{Found: true, Value: value}, nil
}

func (s *CacheServer) Delete(ctx context.Context, req *cache.DeleteRequest) (*cache.DeleteResponse, error) {
	s.cache.Delete(req.Key)
	s.metrics.CacheDeletes.Inc()
	return &cache.DeleteResponse{Success: true}, nil
}

func (s *CacheServer) HealthCheck(ctx context.Context, req *cache.HealthRequest) (*cache.HealthResponse, error) {
	return &cache.HealthResponse{Status: "healthy"}, nil
}

func (s *CacheServer) Start() error {
	listener, err := net.Listen("tcp", s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.port, err)
	}

	s.server = grpc.NewServer()
	cache.RegisterCacheServiceServer(s.server, s)

	log.Printf("Cache server started on %s", s.port)

	return s.server.Serve(listener)
}

func (s *CacheServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
		log.Println("Cache server stopped gracefully")
	}
}
