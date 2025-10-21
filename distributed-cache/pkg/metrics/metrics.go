package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	CacheHits         prometheus.Counter
	CacheMisses       prometheus.Counter
	CacheSets         prometheus.Counter
	CacheDeletes      prometheus.Counter
	RequestDuration   prometheus.Histogram
	ActiveConnections prometheus.Gauge
}

func New() *Metrics {
	return &Metrics{
		CacheHits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		}),
		CacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		}),
		CacheSets: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cache_sets_total",
			Help: "Total number of cache sets",
		}),
		CacheDeletes: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cache_deletes_total",
			Help: "Total number of cache deletes",
		}),
		RequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Request latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		ActiveConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		}),
	}
}
