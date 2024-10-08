package main

import (
	"os"
	"testing"
	"time"
)

func TestLoadQueriesFromFile(t *testing.T) {

	// Create a temporary file with sample SQL queries
	sqlFileContent := "SELECT * FROM users WHERE id = 1;\nSELECT * FROM users WHERE id = 2;\n"
	tmpFile, err := os.CreateTemp("", "queries.sql")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up after the test

	if _, err := tmpFile.Write([]byte(sqlFileContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Load queries from the temporary file
	queries, err := loadQueriesFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load queries from file: %v", err)
	}

	// Verify the loaded queries
	expectedQueries := []string{
		"SELECT * FROM users WHERE id = 1",
		"SELECT * FROM users WHERE id = 2",
	}

	if len(queries) != len(expectedQueries) {
		t.Fatalf("Expected %d queries, got %d", len(expectedQueries), len(queries))
	}

	for i, query := range queries {
		if query != expectedQueries[i] {
			t.Errorf("Expected query: %s, got: %s", expectedQueries[i], query)
		}
	}
}

func TestLoadConfig(t *testing.T) {

	// Create a temporary YAML config file
	yamlContent := `
database:
  dsn: "user:password@tcp(localhost:3306)/dbname?parseTime=true&timeout=5s"
  max_open_conns: 50
  max_idle_conns: 25
  conn_max_lifetime: 60s
  conn_idle_timeout: 30s
  test_query: "SELECT 1"
  query_file: ""
  seed_query: "SELECT id FROM users ORDER BY RAND() LIMIT 5"
  query_template: "SELECT * FROM users WHERE id = ?"
  query_interval: 1s
  concurrent_workers: 5
`

	// Test loading a non-existent config file
	_, err := LoadConfig("foobar-noname-file")
	if err == nil {
		t.Fatalf("Failure expected loading non-existent config file: %v", err)
	}

	// Test failure cases on config
	// Create bad YAML config content
	badContent := `badContent- foo`

	// Create temp file for loading bad content
	badTmpFile, err := os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(badTmpFile.Name()) // Clean up after the test
	if _, err := badTmpFile.Write([]byte(badContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	// Load the bad config file
	cfg, err := LoadConfig(badTmpFile.Name())
	if err == nil {
		t.Fatalf("Failed to verify bad yaml file: %v", err)
	}

	// Create temp file for loading bad content
	tmpFile, err := os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	// Test success cases
	// Write good yaml content
	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Load the config file
	cfg, err = LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate that the config values are correctly loaded
	if cfg.Database.DSN != "user:password@tcp(localhost:3306)/dbname?parseTime=true&timeout=5s" {
		t.Errorf("Unexpected DSN value: %s", cfg.Database.DSN)
	}
	if cfg.Database.MaxOpenConns != 50 {
		t.Errorf("Unexpected MaxOpenConns value: %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Database.ConnMaxLifetime != 60*time.Second {
		t.Errorf("Unexpected ConnMaxLifetime value: %v", cfg.Database.ConnMaxLifetime)
	}
	if cfg.Database.SeedQuery != "SELECT id FROM users ORDER BY RAND() LIMIT 5" {
		t.Errorf("Unexpected SeedQuery value: %s", cfg.Database.SeedQuery)
	}
}
