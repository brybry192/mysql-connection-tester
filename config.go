package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type DatabaseConfig struct {
	DSN               string        `yaml:"dsn"`
	MaxOpenConns      int           `yaml:"max_open_conns"`
	MaxIdleConns      int           `yaml:"max_idle_conns"`
	ConnMaxLifetime   time.Duration `yaml:"conn_max_lifetime"`
	ConnIdleTimeout   time.Duration `yaml:"conn_idle_timeout"`
	TestQuery         string        `yaml:"test_query"`
	QueryFile         string        `yaml:"query_file"`
	SeedQuery         string        `yaml:"seed_query"`
	QueryTemplate     string        `yaml:"query_template"`
	QueryInterval     time.Duration `yaml:"query_interval"`
	ConcurrentWorkers int           `yaml:"concurrent_workers"`
	Queries           []string      `yaml:"queries"`
}

type Config struct {
	Database DatabaseConfig `yaml:"database"`
}

// LoadConfig loads configuration from yaml and environment variables
func LoadConfig(configFile string) (*Config, error) {
	var cfg Config

	// Read YAML config
	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}

	// Override with environment variables if present
	viper.SetEnvPrefix("MYSQLTESTER")
	viper.AutomaticEnv()

	viper.SetDefault("DATABASE_DSN", cfg.Database.DSN)
	viper.SetDefault("DATABASE_MAX_OPEN_CONNS", cfg.Database.MaxOpenConns)
	viper.SetDefault("DATABASE_MAX_IDLE_CONNS", cfg.Database.MaxIdleConns)
	viper.SetDefault("DATABASE_CONN_MAX_LIFETIME", cfg.Database.ConnMaxLifetime)
	viper.SetDefault("DATABASE_CONN_IDLE_TIMEOUT", cfg.Database.ConnIdleTimeout)
	viper.SetDefault("DATABASE_TEST_QUERY", cfg.Database.TestQuery)
	viper.SetDefault("DATABASE_QUERY_FILE", cfg.Database.QueryFile)
	viper.SetDefault("DATABASE_SEED_QUERY", cfg.Database.SeedQuery)
	viper.SetDefault("DATABASE_QUERY_TEMPLATE", cfg.Database.QueryTemplate)
	viper.SetDefault("DATABASE_QUERY_INTERVAL", cfg.Database.QueryInterval)
	viper.SetDefault("DATABASE_CONCURRENT_WORKERS", cfg.Database.ConcurrentWorkers)

	cfg.Database.DSN = viper.GetString("DATABASE_DSN")
	cfg.Database.MaxOpenConns = viper.GetInt("DATABASE_MAX_OPEN_CONNS")
	cfg.Database.MaxIdleConns = viper.GetInt("DATABASE_MAX_IDLE_CONNS")
	cfg.Database.ConnMaxLifetime = viper.GetDuration("DATABASE_CONN_MAX_LIFETIME")
	cfg.Database.ConnIdleTimeout = viper.GetDuration("DATABASE_CONN_IDLE_TIMEOUT")
	cfg.Database.TestQuery = viper.GetString("DATABASE_TEST_QUERY")
	cfg.Database.QueryFile = viper.GetString("DATABASE_QUERY_FILE")
	cfg.Database.SeedQuery = viper.GetString("DATABASE_SEED_QUERY")
	cfg.Database.QueryTemplate = viper.GetString("DATABASE_QUERY_TEMPLATE")
	cfg.Database.QueryInterval = viper.GetDuration("DATABASE_QUERY_INTERVAL")
	cfg.Database.ConcurrentWorkers = viper.GetInt("DATABASE_CONCURRENT_WORKERS")

	return &cfg, nil
}

// loadQueriesFromFile reads and splits the SQL queries from the file
func loadQueriesFromFile(filePath string) ([]string, error) {
	// Read the entire file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading query file: %w", err)
	}

	// Split the content into separate queries using ";" as a delimiter
	queries := strings.Split(string(content), ";")

	// Clean up each query by trimming whitespace and filtering out empty entries
	var cleanQueries []string
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query != "" {
			cleanQueries = append(cleanQueries, query)
		}
	}
	return cleanQueries, nil
}
