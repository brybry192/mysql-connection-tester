package main

import (
	"log"
)

func main() {
	// Customize log output format
	log.SetFlags(0) // Disable default timestamp
	log.SetOutput(new(logWriter))

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
