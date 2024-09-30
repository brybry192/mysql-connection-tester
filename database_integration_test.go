package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Global variables for shared container and database connection
var (
	mysqlContainer testcontainers.Container
	MysqlHost      string = "127.0.0.1" // To store the MySQL container host
	MysqlPort      string = "13306"     // To store the MySQL container port
)

// TestMain handles setup and teardown for all integration tests
func TestMain(m *testing.M) {
	var err error
	// Setup MySQL container before running tests
	mysqlContainer, db, err = setupMySQLContainer()
	if err != nil {
		log.Fatalf("Failed to set up MySQL container: %v", err)
	}

	// Run the tests
	code := m.Run()

	// Teardown container after tests complete
	if mysqlContainer != nil {
		if err := mysqlContainer.Terminate(context.Background()); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}

	os.Exit(code)
}

// waitForDatabaseConnection ensures the database is ready before running the tests
func waitForDatabaseConnection(db *sqlx.DB, timeout time.Duration) error {
	start := time.Now()
	for {
		err := db.Ping()
		if err == nil {
			return nil
		}

		if time.Since(start) > timeout {
			return fmt.Errorf("timeout reached while waiting for database connection: %v", err)
		}

		time.Sleep(1 * time.Second) // Sleep briefly before retrying
	}
}

// setupMySQLContainer starts a single MySQL container for all integration tests
func setupMySQLContainer() (testcontainers.Container, *sqlx.DB, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mysql:8.0",
		Env:          map[string]string{"MYSQL_ROOT_PASSWORD": "password", "MYSQL_DATABASE": "testdb"},
		ExposedPorts: []string{"3306/tcp"},
		WaitingFor:   wait.ForListeningPort("3306/tcp").WithStartupTimeout(30 * time.Second),
	}

	mysqlContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start MySQL container: %w", err)
	}

	// Get the container's host and mapped port
	MysqlHost, err = mysqlContainer.Host(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get MySQL container host: %w", err)
	}

	// Fetch the explicitly mapped port 13306
	mappedPort, err := mysqlContainer.MappedPort(ctx, "3306")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get MySQL container port: %w", err)
	}
	MysqlPort = mappedPort.Port()

	dsn := fmt.Sprintf("root:password@tcp(%s:%s)/testdb?parseTime=true", MysqlHost, MysqlPort)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to MySQL container: %w", err)
	}

	// Wait until the connection is ready
	for i := 0; i < 20; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
		if i == 9 {
			return nil, nil, fmt.Errorf("MySQL container not ready: %w", err)
		}
	}

	// Create the database
	_, err = db.Exec(`show databases`)

	// Set up the table and test data
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id BIGINT NOT NULL AUTO_INCREMENT, user VARCHAR(255) DEFAULT NULL, name VARCHAR(255) DEFAULT NULL, PRIMARY KEY (id))`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO users (user, name) VALUES ('foobar', 'Foo Bar')`)
	if err != nil {
		log.Fatalf("Failed to insert test data: %v", err)
	}

	return mysqlContainer, db, nil
}

// TestInitializeDB tests the InitializeDB function
func TestInitializeDB(t *testing.T) {
	// Define configuration for testing
	cfg := &Config{
		Debug: true,
		Database: DatabaseConfig{
			DSN:               fmt.Sprintf("root:password@tcp(%s:%s)/testdb?parseTime=true", MysqlHost, MysqlPort),
			MaxOpenConns:      10,
			MaxIdleConns:      5,
			ConnMaxLifetime:   30 * time.Second,
			ConnIdleTimeout:   15 * time.Second,
			ConcurrentWorkers: 1,                                  // Use more than one worker for better coverage
			QueriesPerWorker:  1,                                  // Ensure multiple queries per worker
			SeedQuery:         "SELECT id FROM users LIMIT 5",     // Provide a valid seed query
			QueryTemplate:     "SELECT * FROM users WHERE id = ?", // Valid query template
		},
	}

	// Test InitializeDB function
	dbWrapper, err := InitializeDBWrapper(cfg)
	if err != nil {

		t.Fatalf("InitializeDB failed: %v", err)
	}
	defer dbWrapper.Close()

	// Check connection parameters
	if dbWrapper.DB.Stats().MaxOpenConnections != 10 {
		t.Errorf("Expected MaxOpenConnections to be 10, got %d", dbWrapper.DB.Stats().MaxOpenConnections)
	}
}
