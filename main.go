package main

import (
	"fmt"
	"os"

	"github.com/rishiyaduwanshi/boiler/internal/cli"
	"github.com/rishiyaduwanshi/boiler/internal/config"
	"github.com/rishiyaduwanshi/boiler/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize directories
	if err := cfg.InitializeDirs(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing directories: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	logger, err := utils.NewLogger(cfg.Paths.Logs, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Boiler started")

	// Execute CLI
	if err := cli.Execute(cfg, logger); err != nil {
		logger.Error(fmt.Sprintf("Execution error: %v", err))
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}