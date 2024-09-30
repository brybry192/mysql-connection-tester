package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/rand"
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
	db.SetConnMaxIdleTime(cfg.Database.ConnIdleTimeout) // This must be set before MaxLifeTime
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return &DBWrapper{
		DB:    db,
		Close: func() { db.Close() },
	}, nil
}

// RunQueryWorkers runs multiple test queries in parallel within a single worker
func RunQueryWorkers(cfg *Config, db *sqlx.DB, workerID int) {
	numQueriesPerWorker := cfg.Database.QueriesPerWorker // Number of concurrent queries per worker
	ticker := time.NewTicker(cfg.Database.QueryInterval)
	defer ticker.Stop()

	// Execute the seed query to fetch input values
	_, inputValues, err := genericQuery(db, cfg.Database.SeedQuery, nil)
	if err != nil || len(inputValues) == 0 {
		queryErrors.WithLabelValues(fmt.Sprintf("%d", workerID), "worker_id").Inc()
		log.Printf("[Worker %d] Failed to fetch seed values: %v", workerID, err)
		return
	}

	// Warm up the connection pool
	warmUpConnections(db, cfg)

	for i := 0; i < numQueriesPerWorker; i++ {

		log.Printf("Starting [Worker %d - Query %d]", workerID, i)
		for {
			select {
			case <-ticker.C:
				// Get a random index
				randomIndex := rand.Intn(len(inputValues))
				seedRow := inputValues[randomIndex]
				// Prepare the value slice for the query execution from the seedData
				var queryValues []interface{}
				for _, value := range seedRow {
					queryValues = append(queryValues, value)
				}

				startTime := time.Now() // Start time tracking
				// Execute the query template with the seed values
				_, rows, err := genericQuery(db, cfg.Database.QueryTemplate, queryValues)
				duration := time.Since(startTime).Seconds() // Calculate duration
				queryDuration.WithLabelValues(fmt.Sprintf("%d", workerID), "worker_id").Observe(duration)

				if err != nil {
					queryErrors.WithLabelValues(fmt.Sprintf("%d", workerID), "worker_id").Inc()
					log.Printf("[Worker %d - Query %d] Query failed: %v\n", workerID, i, err)
					continue
				}
				if debug {
					log.Printf("[Worker %d - Query %d] Executed query: %v", workerID, i, rows)
				}
			}
		}

	}
}

// executeQueryWithValues runs the query template with the provided values
func executeQueryWithValues(db *sqlx.DB, queryTemplate string, values []interface{}) error {
	rows, err := db.Queryx(queryTemplate, values...)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Printf("%v+", db)
	for rows.Next() {

		columns, err := rows.SliceScan()
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("Query returned no rows")
				return nil
			}
			return fmt.Errorf("error executing query: %w", err)
		}

		// Process each column value and convert as necessary
		convertedValues := make([]interface{}, len(columns))
		for i, col := range columns {
			convertedValues[i] = convertValue(col)
		}

		// Log the converted values
		log.Printf("Query worker %v response: %v", convertedValues[0], convertedValues[1])
	}

	return nil
}

// convertValue converts raw query results to a more readable format
func convertValue(val interface{}) interface{} {
	switch v := val.(type) {
	case []byte:
		// Convert []byte to string
		return string(v)
	case nil:
		// Handle null values gracefully
		return "NULL"
	default:
		// Return the value as-is if it's already in a suitable format
		return v
	}
}

// genericQuery runs a query and returns columns and rows with their proper types
func genericQuery(db *sqlx.DB, query string, values []interface{}) ([]string, []map[string]interface{}, error) {
	// Execute the query
	rows, err := db.Queryx(query, values...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	// Prepare a slice to hold the result set
	var result []map[string]interface{}

	// Create a slice of interface{}'s to hold each row's column values
	// and a slice of pointers to each value for scanning
	for rows.Next() {
		columnPointers := make([]interface{}, len(columns))
		columnValues := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		// Scan the row into columnPointers
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, nil, err
		}

		// Create a map to store the column name and value for each row
		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := columnValues[i]
			// Check if the value is a byte slice (usually for strings or blobs)
			if b, ok := val.([]byte); ok {
				rowMap[colName] = string(b)
			} else {
				rowMap[colName] = convertToProperType(val)
			}
		}

		// Append the row to the result set
		result = append(result, rowMap)
	}

	// Check for any error encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return columns, result, nil
}

// convertToProperType converts SQL types to Go's types
func convertToProperType(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case int64, float64, bool, string:
		return v
	default:
		// Use reflection to handle any other types that might occur
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr && !rv.IsNil() {
			return rv.Elem().Interface()
		}
		return value
	}
}

// warmUpConnections performs simple queries to establish idle connections
func warmUpConnections(db *sqlx.DB, cfg *Config) error {

	//db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnIdleTimeout)
	//db.SetConnMaxIdleTime(7 * time.Second)
	for i := 0; i < cfg.Database.NumIdleConnections; i++ {
		log.Printf("Warmed up connection: %d", i+1)
		go func() {

			var test int
			err := db.Get(&test, "SELECT 1")
			if err != nil {
				log.Printf("Error warming up connection: %v", err)
				return
			}
			log.Printf("Warmed up connection: %d", i+1)
		}()
	}
	return nil
}

func collectDBPoolMetrics(db *sqlx.DB, poolName string, interval time.Duration) {
	for {
		stats := db.Stats()
		openConnections.WithLabelValues(poolName).Set(float64(stats.OpenConnections))
		idleConnections.WithLabelValues(poolName).Set(float64(stats.Idle))
		inUseConnections.WithLabelValues(poolName).Set(float64(stats.InUse))
		time.Sleep(interval) // Collect metrics every time duration
	}
	return
}
