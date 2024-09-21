package cmd

import (
	"fmt"
	"testing"
	"time"

	"mysql-connection-tester/config"
	"mysql-connection-tester/database"
)

// TestStartCmdIntegration uses the shared MySQL container from the database package
func TestStartCmdIntegration(t *testing.T) {

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DSN:             fmt.Sprintf("root:password@tcp(%s:%s)/testdb?parseTime=true", database.MysqlHost, database.MysqlPort),
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 30 * time.Second,
			ConnIdleTimeout: 15 * time.Second,
			TestQuery:       "SELECT 1",
			QueryInterval:   1 * time.Second,
			ConcurrentWorkers: 1,
		},
	}

	// Start `StartCmdWithConfig` using the shared DB connection
	if err := StartCmdWithConfig(cfg, func(cfg *config.Config) (*database.DBWrapper, error) {
		return database.InitializeDBWrapper(cfg)
	}); err != nil {
		t.Fatalf("StartCmdWithConfig failed: %v", err)
	}

	// Ensure the test executed successfully
	t.Logf("Integration test for StartCmdWithConfig completed successfully")
}
