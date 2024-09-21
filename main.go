package main

import (
	"log"
)

func main() {
	// Load configuration
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Start the command using the loaded configuration and the real database initializer
	if err := StartCmdWithConfig(cfg, InitializeDBWrapper); err != nil {
		log.Fatalf("Application failed to start: %v", err)
	}
}
