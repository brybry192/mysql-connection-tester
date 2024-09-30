package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define Prometheus metrics
var (
	queryErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_errors_total",
			Help: "Total number of SQL query errors",
		},
		[]string{"worker_id", "query"},
	)

	queryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Histogram of SQL query execution times",
			Buckets: prometheus.DefBuckets, // Default buckets: [0.005, 0.01, 0.025, ...]
		},
		[]string{"worker_id", "query"},
	)

	openConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_open_connections",
			Help: "Number of open connections in the DB connection pool",
		},
		[]string{"pool"},
	)

	idleConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_idle_connections",
			Help: "Number of idle connections in the DB connection pool",
		},
		[]string{"pool"},
	)

	inUseConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_in_use_connections",
			Help: "Number of in-use connections in the DB connection pool",
		},
		[]string{"pool"},
	)
)

// Register metrics with prometheus
func init() {
	prometheus.MustRegister(queryErrors)
	prometheus.MustRegister(queryDuration)
	prometheus.MustRegister(openConnections)
	prometheus.MustRegister(idleConnections)
	prometheus.MustRegister(inUseConnections)
}

// This application isn't a web app, so start dedicated http server for prometheus
func startMetricsServer(port, metricsPath string) {
	http.Handle(metricsPath, promhttp.Handler())
	log.Printf("Starting prometheus server on :%s/metrics\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
