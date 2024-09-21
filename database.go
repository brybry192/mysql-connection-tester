package main

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Global variables for shared container and database connection
var (
	mysqlContainer testcontainers.Container
	db             *sqlx.DB
	MysqlHost      string
	MysqlPort      string
)

// DBWrapper is a wrapper for handling the database connection and the mock
type DBWrapper struct {
	DB    *sqlx.DB
	Close func()
}

// InitializeDBWrapper initializes the DB connection and sets the appropriate configurations
func InitializeDBWrapper(cfg *Config) (*DBWrapper, error) {
	db, err := sqlx.Connect("mysql", cfg.Database.DSN)
	if err != nil {
		return nil, err
	}

	// Set the connection pool parameters
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return &DBWrapper{
		DB:    db,
		Close: func() { db.Close() },
	}, nil
}

// TestLoop runs multiple test queries in parallel within a single worker
func TestLoop(cfg *Config, db *sqlx.DB, workerID int) {
	// Your existing TestLoop logic...
}

// setupMySQLContainer starts a single MySQL container for all integration tests
func setupMySQLContainer() error {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mysql:8.0",
		Env:          map[string]string{"MYSQL_ROOT_PASSWORD": "password", "MYSQL_DATABASE": "testdb"},
		ExposedPorts: []string{"3306/tcp"},
		WaitingFor:   wait.ForListeningPort("3306/tcp").WithStartupTimeout(2 * time.Minute),
	}

	var err error
	mysqlContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return fmt.Errorf("failed to start MySQL container: %w", err)
	}

	MysqlHost, err = mysqlContainer.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get MySQL container host: %w", err)
	}

	mappedPort, err := mysqlContainer.MappedPort(ctx, "3306")
	if err != nil {
		return fmt.Errorf("failed to get MySQL container port: %w", err)
	}
	MysqlPort = mappedPort.Port()

	return nil
}
