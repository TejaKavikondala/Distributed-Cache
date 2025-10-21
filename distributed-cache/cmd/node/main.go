package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/TejaKavikondala/Distributed-Cache/pkg/metrics"
	"github.com/TejaKavikondala/Distributed-Cache/pkg/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	port := flag.String("port", ":50051", "Server port")
	metricsPort := flag.String("metrics-port", ":9090", "Prometheus metrics port")
	flag.Parse()

	m := metrics.New()

	// Allow overriding via environment variables commonly used in Docker
	if env := os.Getenv("PORT"); env != "" {
		*port = env
	}
	if env := os.Getenv("METRICS_PORT"); env != "" {
		*metricsPort = env
	}

	go func() {
		log.Printf("Metrics server on http://localhost%s/metrics", *metricsPort)
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(*metricsPort, nil))
	}()

	srv := server.New(*port, m)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gracefully...")
		srv.Stop()
		os.Exit(0)
	}()

	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
