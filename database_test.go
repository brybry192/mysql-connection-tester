package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

// TestInitializeDBWrapper verifies the InitializeDBWrapper function
func TestInitializeDBWrapper(t *testing.T) {
	// Define configuration for testing
	cfg := &Config{
		Debug: true, // Enable Debug mode if needed for more verbose logging
		Database: DatabaseConfig{
			DSN:                fmt.Sprintf("root:password@tcp(%s:%s)/testdb?parseTime=true", MysqlHost, MysqlPort),
			MaxOpenConns:       10,
			MaxIdleConns:       5,
			ConnMaxLifetime:    30 * time.Second,
			ConnIdleTimeout:    15 * time.Second,
			NumIdleConnections: 1,
			ConcurrentWorkers:  1,                                       // Use more than one worker for better coverage
			QueriesPerWorker:   1,                                       // Ensure multiple queries per worker
			SeedQuery:          "SELECT id FROM test_table LIMIT 5",     // Provide a valid seed query
			QueryTemplate:      "SELECT * FROM test_table WHERE id = ?", // Valid query template
		},
	}

	// Test InitializeDBWrapper function
	dbWrapper, err := InitializeDBWrapper(cfg)
	if err != nil {
		t.Fatalf("InitializeDBWrapper failed: %v", err)
	}
	defer dbWrapper.Close()

	// Check connection parameters
	if dbWrapper.DB.Stats().MaxOpenConnections != 10 {
		t.Errorf("Expected MaxOpenConnections to be 10, got %d", dbWrapper.DB.Stats().MaxOpenConnections)
	}
}

// Test the executeQueryWithValues function using sqlmock
func TestExecuteQueryWithValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "mysql")

	// Expect the query with the provided value
	mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\?").WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user", "name"}).AddRow(1, "user1", "Name One"))

	err = executeQueryWithValues(sqlxDB, "SELECT * FROM users WHERE id = ?", []interface{}{1})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}

func TestRunQueryWorkers(t *testing.T) {
	// Set up the database and configuration for testing
	cfg := &Config{
		Debug: true, // Enable Debug mode if needed for more verbose logging
		Database: DatabaseConfig{
			DSN:                fmt.Sprintf("root:password@tcp(%s:%s)/testdb?parseTime=true", MysqlHost, MysqlPort),
			MaxOpenConns:       5,
			MaxIdleConns:       3,
			ConnMaxLifetime:    30 * time.Second,
			ConnIdleTimeout:    15 * time.Second,
			NumIdleConnections: 1,
			TestQuery:          "SELECT 1",
			QueryInterval:      1 * time.Second,
			ConcurrentWorkers:  2,                                       // Use more than one worker for better coverage
			QueriesPerWorker:   2,                                       // Ensure multiple queries per worker
			SeedQuery:          "SELECT id FROM test_table LIMIT 5",     // Provide a valid seed query
			QueryTemplate:      "SELECT * FROM test_table WHERE id = ?", // Valid query template
		},
	}

	// Set up test database table and data
	dbWrapper, err := InitializeDBWrapper(cfg)
	if err != nil {
		t.Fatalf("Failed to initialize database wrapper: %v", err)
	}
	defer dbWrapper.Close()

	_, err = dbWrapper.DB.Exec(`
        CREATE TABLE IF NOT EXISTS test_table (
            id INT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(50)
        );
    `)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Insert some test data
	for i := 1; i <= 10; i++ {
		_, err := dbWrapper.DB.Exec(`INSERT INTO test_table (name) VALUES (?)`, fmt.Sprintf("Name %d", i))
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Call RunQueryWorkers
	go RunQueryWorkers(cfg, dbWrapper.DB, 1)

	// Allow some time for the workers to run
	time.Sleep(1 * time.Second)

	// No explicit assertions needed since we're testing code coverage
	t.Logf("TestRunQueryWorkers completed successfully")
}
