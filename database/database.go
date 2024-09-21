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

// TestLoop runs the test query in an interval using a specific connection pool
func TestLoop(cfg *config.Config, db *sqlx.DB, workerID int) {
	ticker := time.NewTicker(cfg.Database.QueryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var result int
			err := db.QueryRow(cfg.Database.TestQuery).Scan(&result)
			if err != nil {
				if err == sql.ErrNoRows {
					log.Printf("[Worker %d] Query returned no rows", workerID)
				} else {
					log.Printf("[Worker %d] Query failed: %v", workerID, err)
				}
			} else {
				log.Printf("[Worker %d] Query succeeded, result: %d", workerID, result)
			}
		}
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
