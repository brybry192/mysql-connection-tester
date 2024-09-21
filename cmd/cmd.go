package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"mysql-connection-tester/config"
	"mysql-connection-tester/database"
)

func StartCmd() {
	// Load the configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Start multiple workers based on the configuration
	for i := 0; i < cfg.Database.ConcurrentWorkers; i++ {
		// Initialize a separate connection pool for each worker
		db, err := database.InitializeDB(cfg)
		if err != nil {
			log.Fatalf("Worker %d: Failed to initialize database: %v", i, err)
		}

		log.Printf("Worker %d: Connected to the database successfully", i)

		// Start a TestLoop in a separate goroutine for each worker
		go database.TestLoop(cfg, db, i)
	}

	// Set up signal handling to allow graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	log.Println("Shutting down gracefully")
}
