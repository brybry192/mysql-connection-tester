package main

import (
	"fmt"
	"testing"
	"time"
)

func TestStartCmdIntegration(t *testing.T) {
	// Set up a channel to stop the test after a specified period
	stopChan := make(chan struct{})

	// Start a goroutine that will close the stop channel after 10 seconds
	go func() {
		time.Sleep(10 * time.Second)
		close(stopChan)
	}()

	cfg := &Config{
		Database: DatabaseConfig{
			DSN:               fmt.Sprintf("root:password@tcp(%s:%s)/testdb?parseTime=true", MysqlHost, MysqlPort),
			MaxOpenConns:      5,
			MaxIdleConns:      3,
			ConnMaxLifetime:   30 * time.Second,
			ConnIdleTimeout:   15 * time.Second,
			TestQuery:         "SELECT 1",
			QueryInterval:     1 * time.Second,
			ConcurrentWorkers: 1,
		},
	}

	// Start `StartCmdWithConfig` using the shared DB connection in a separate goroutine
	go func() {
		if err := StartCmdWithConfig(cfg, InitializeDBWrapper); err != nil {
			t.Fatalf("StartCmdWithConfig failed: %v", err)
		}
	}()

	// Wait for the stopChan to be closed, indicating the end of the test duration
	<-stopChan

	t.Logf("Integration test for StartCmdWithConfig completed successfully")
}
