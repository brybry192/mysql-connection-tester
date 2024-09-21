package main

import (
	"log"
	"mysql-connection-tester/cmd"
	"mysql-connection-tester/config"
	"mysql-connection-tester/database"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Start the command using the loaded configuration and the real database initializer
	if err := cmd.StartCmdWithConfig(cfg, database.InitializeDBWrapper); err != nil {
		log.Fatalf("Application failed to start: %v", err)
	}
}
