package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

// Ensure you reset metrics before running tests
func resetMetrics() {
	queryErrors.Reset()
	queryDuration.Reset()
	openConnections.Reset()
	idleConnections.Reset()
	inUseConnections.Reset()
}

func TestQueryErrorsMetric(t *testing.T) {
	resetMetrics() // Reset metrics at the beginning of the test

	// Simulate an error
	workerID := "1"
	query := "test_query"
	queryErrors.WithLabelValues(workerID, query).Inc()

	// Check that the metric value is incremented correctly
	metricValue := testutil.ToFloat64(queryErrors.WithLabelValues(workerID, query))
	if metricValue != 1 {
		t.Errorf("Expected queryErrors metric to be 1, got %v", metricValue)
	}
}

func TestQueryDurationMetric(t *testing.T) {
	resetMetrics() // Reset metrics at the beginning of the test

	// Simulate recording a query duration
	workerID := "2"
	query := "test_duration_query"
	queryDuration.WithLabelValues(workerID, query).Observe(2.5)

	// Collect metrics for verification
	collected := testutil.CollectAndCount(queryDuration, "db_query_duration_seconds")
	if collected == 0 {
		t.Errorf("Expected db_query_duration_seconds to be collected")
	}
}

func TestMetricsEndpoint(t *testing.T) {
	resetMetrics() // Reset metrics before starting the test

	// Increment the error metric to make sure it's present
	queryErrors.WithLabelValues("1", "test_query").Inc()

	// Give the server some time to start
	time.Sleep(1 * time.Second)

	// Make an HTTP request to the metrics endpoint
	resp, err := http.Get("http://localhost:2112/metrics")
	if err != nil {
		t.Fatalf("Error fetching metrics endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Check if response code is 200 (OK)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %v", resp.StatusCode)
	}

	// Optionally, you can read the response body and check for specific metrics
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "db_query_errors_total") {
		t.Errorf("Expected db_query_errors_total metric to be present")
	}
}
