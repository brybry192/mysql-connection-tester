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
		Database: DatabaseConfig{
			DSN:             fmt.Sprintf("root:password@tcp(%s:%s)/testdb?parseTime=true", MysqlHost, MysqlPort),
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 30 * time.Second,
			ConnIdleTimeout: 15 * time.Second,
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
