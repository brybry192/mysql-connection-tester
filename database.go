package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Global variables for shared container and database connection
var (
	db *sqlx.DB
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
	db.SetConnMaxIdleTime(cfg.Database.ConnIdleTimeout)

	return &DBWrapper{
		DB:    db,
		Close: func() { db.Close() },
	}, nil
}

// TestLoop runs multiple test queries in parallel within a single worker
func TestLoop(cfg *Config, db *sqlx.DB, workerID int) {
	// Your existing TestLoop logic...
}

// executeQueryWithValues runs the query template with the provided values
func executeQueryWithValues(db *sqlx.DB, queryTemplate string, values []interface{}) error {
	row := db.QueryRowx(queryTemplate, values...)
	columns, err := row.SliceScan()
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Query returned no rows")
			return nil
		}
		return fmt.Errorf("error executing query: %w", err)
	}

	log.Printf("Query succeeded with values: %v", columns)
	return nil
}
