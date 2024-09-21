package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		DSN             string        `yaml:"dsn"`
		MaxOpenConns    int           `yaml:"max_open_conns"`
		MaxIdleConns    int           `yaml:"max_idle_conns"`
		ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
		ConnIdleTimeout time.Duration `yaml:"conn_idle_timeout"`
		TestQuery       string        `yaml:"test_query"`
		QueryInterval   time.Duration `yaml:"query_interval"`
		ConcurrentWorkers int         `yaml:"concurrent_workers"` // New field for the number of workers
	} `yaml:"database"`
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
	viper.SetDefault("DATABASE_QUERY_INTERVAL", cfg.Database.QueryInterval)
	viper.SetDefault("DATABASE_CONCURRENT_WORKERS", cfg.Database.ConcurrentWorkers)

	cfg.Database.DSN = viper.GetString("DATABASE_DSN")
	cfg.Database.MaxOpenConns = viper.GetInt("DATABASE_MAX_OPEN_CONNS")
	cfg.Database.MaxIdleConns = viper.GetInt("DATABASE_MAX_IDLE_CONNS")
	cfg.Database.ConnMaxLifetime = viper.GetDuration("DATABASE_CONN_MAX_LIFETIME")
	cfg.Database.ConnIdleTimeout = viper.GetDuration("DATABASE_CONN_IDLE_TIMEOUT")
	cfg.Database.TestQuery = viper.GetString("DATABASE_TEST_QUERY")
	cfg.Database.QueryInterval = viper.GetDuration("DATABASE_QUERY_INTERVAL")
	cfg.Database.ConcurrentWorkers = viper.GetInt("DATABASE_CONCURRENT_WORKERS")

	return &cfg, nil
}
