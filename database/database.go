package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"mysql-connection-tester/config"
)

// TestLoop runs multiple test queries in parallel within a single worker
func TestLoop(cfg *config.Config, db *sqlx.DB, workerID int) {
	numQueriesPerWorker := 1 // Number of concurrent queries per worker
	ticker := time.NewTicker(cfg.Database.QueryInterval / time.Duration(numQueriesPerWorker))
	defer ticker.Stop()

	for i := 0; i < numQueriesPerWorker; i++ {
		go func(queryIndex int) {
			// Execute the seed query to fetch input values
                        inputValues, err := fetchSeedValues(db, cfg.Database.SeedQuery)
			if err != nil || len(inputValues) == 0 {
				log.Printf("[Worker %d - Query %d] Failed to fetch seed values: %v", workerID, queryIndex, err)
				return
			}
			n := 0
			for {
				select {
				case <-ticker.C:
					log.Printf("[Worker %d - Query %d] Executing query template with value: %v", workerID, queryIndex, inputValues[n])
					// Execute the query template with the seed values
					if err := executeQueryWithValues(db, cfg.Database.QueryTemplate, inputValues[n]); err != nil {
						log.Printf("[Worker %d - Query %d] Query failed: %v", workerID, queryIndex, err)
					}
				}
				if n < len(inputValues) {
					n = n + 1
				} else if n == len(inputValues) {
					n = 0
				}
			}
		}(i)
	}
}

// InitializeDB initializes a database connection pool
func InitializeDB(cfg *config.Config) (*sqlx.DB, error) {
        db, err := sqlx.Connect("mysql", cfg.Database.DSN)
        if err != nil {
                return nil, fmt.Errorf("failed to connect to the database: %w", err)
        }

        db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
        db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
        db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
        db.SetConnMaxIdleTime(cfg.Database.ConnIdleTimeout)

        return db, nil
}

// fetchSeedValues executes the seed query and returns the result set
func fetchSeedValues(db *sqlx.DB, seedQuery string) ([][]interface{}, error) {
	rows, err := db.Queryx(seedQuery)
	if err != nil {
		return nil, fmt.Errorf("error executing seed query: %w", err)
	}
	defer rows.Close()

	var results [][]interface{}
	for rows.Next() {
		columns, err := rows.SliceScan()
		if err != nil {
			return nil, fmt.Errorf("error scanning seed query result: %w", err)
		}
		results = append(results, columns)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating seed query results: %w", err)
	}

	return results, nil
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
