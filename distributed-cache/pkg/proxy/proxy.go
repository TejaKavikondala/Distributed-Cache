package proxy

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	cache "github.com/TejaKavikondala/Distributed-Cache/pkg/cache"
	"github.com/TejaKavikondala/Distributed-Cache/pkg/metrics"
	"github.com/TejaKavikondala/Distributed-Cache/pkg/ring"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Proxy struct {
	ring       *ring.Ring
	connPool   map[string]*grpc.ClientConn
	poolMu     sync.RWMutex
	metrics    *metrics.Metrics
	defaultTTL time.Duration
}

func New(nodes []string, replicas int, m *metrics.Metrics) *Proxy {
	r := ring.New(replicas)
	for _, node := range nodes {
		r.AddNode(node)
	}

	return &Proxy{
		ring:       r,
		connPool:   make(map[string]*grpc.ClientConn),
		metrics:    m,
		defaultTTL: 24 * time.Hour,
	}
}

func (p *Proxy) getConn(addr string) (cache.CacheServiceClient, error) {
	p.poolMu.RLock()
	conn, ok := p.connPool[addr]
	p.poolMu.RUnlock()

	if !ok {
		var err error
		conn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}

		p.poolMu.Lock()
		p.connPool[addr] = conn
		p.poolMu.Unlock()
		p.metrics.ActiveConnections.Inc()
	}

	return cache.NewCacheServiceClient(conn), nil
}

func (p *Proxy) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	node := p.ring.GetNode(key)
	if node == "" {
		return fmt.Errorf("no cache nodes available")
	}

	tl := ttl
	if tl == 0 {
		tl = p.defaultTTL
	}

	client, err := p.getConn(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", node, err)
	}

	_, err = client.Set(ctx, &cache.SetRequest{
		Key:        key,
		Value:      value,
		TtlSeconds: int32(tl.Seconds()),
	})
	return err
}

func (p *Proxy) Get(ctx context.Context, key string) ([]byte, error) {
	node := p.ring.GetNode(key)
	if node == "" {
		return nil, fmt.Errorf("no cache nodes available")
	}

	client, err := p.getConn(node)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", node, err)
	}

	resp, err := client.Get(ctx, &cache.GetRequest{Key: key})
	if err != nil {
		return nil, err
	}

	if !resp.Found {
		return nil, fmt.Errorf("key not found")
	}

	return resp.Value, nil
}

func (p *Proxy) Delete(ctx context.Context, key string) error {
	node := p.ring.GetNode(key)
	if node == "" {
		return fmt.Errorf("no cache nodes available")
	}

	client, err := p.getConn(node)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", node, err)
	}

	_, err = client.Delete(ctx, &cache.DeleteRequest{Key: key})
	return err
}

func (p *Proxy) AddNode(addr string) {
	p.ring.AddNode(addr)
	log.Printf("Added node: %s", addr)
}

func (p *Proxy) RemoveNode(addr string) {
	p.ring.RemoveNode(addr)

	p.poolMu.Lock()
	if conn, ok := p.connPool[addr]; ok {
		_ = conn.Close()
		delete(p.connPool, addr)
		p.metrics.ActiveConnections.Dec()
	}
	p.poolMu.Unlock()

	log.Printf("Removed node: %s", addr)
}

func (p *Proxy) Close() {
	p.poolMu.Lock()
	defer p.poolMu.Unlock()

	for _, conn := range p.connPool {
		_ = conn.Close()
	}
	p.connPool = make(map[string]*grpc.ClientConn)
}
