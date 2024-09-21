package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// StartCmdWithConfig allows for starting with dependency injection (for testing)
func StartCmdWithConfig(cfg *Config, dbInitFunc func(cfg *Config) (*DBWrapper, error)) error {
	// Initialize the database connection using the injected function
	dbWrapper, err := dbInitFunc(cfg)
	if err != nil {
		return err
	}
	defer dbWrapper.Close()
	log.Println("Connected to the database successfully")

	// Start multiple workers based on the configuration
	for i := 0; i < cfg.Database.ConcurrentWorkers; i++ {
		go TestLoop(cfg, dbWrapper.DB, i)
	}

	// Set up signal handling to allow graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("Shutting down gracefully")
	return nil
}
