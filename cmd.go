package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Global for debug logging
var debug bool

// logWriter implements io.Writer
type logWriter struct{}

func (lg *logWriter) Write(p []byte) (int, error) {
	// Format the current time with dashes and customize
	return fmt.Printf("%v %v", time.Now().Format("2006-01-02 15:04:05"), string(p))
}

// StartCmdWithConfig allows for starting with dependency injection (for testing)
func StartCmdWithConfig(cfg *Config, dbInitFunc func(cfg *Config) (*DBWrapper, error)) error {

	// Customize log output format
	log.SetFlags(0) // Disable default timestamp
	log.SetOutput(new(logWriter))

	// Initialize the database connection using the injected function
	dbWrapper, err := dbInitFunc(cfg)
	if err != nil {
		return err
	}
	defer dbWrapper.Close()
	log.Println("Connected to the database successfully")

	// Start prometheus server
	go startMetricsServer(cfg.MetricsPort, "/metrics")

	// Start multiple workers based on the configuration
	for i := 0; i < cfg.Database.ConcurrentWorkers; i++ {
		go RunQueryWorkers(cfg, dbWrapper.DB, i)
		go collectDBPoolMetrics(dbWrapper.DB, fmt.Sprintf("%d", i), cfg.MetricsInterval)
	}

	// Set up signal handling to allow graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("Shutting down gracefully")
	return nil
}
