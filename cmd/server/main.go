// File: cmd/server/main.go

package main

import (
	"basic-go-redis/internal/server"
	"basic-go-redis/pkg/config"
	"basic-go-redis/pkg/logger"
	"flag"
	"fmt"
	"os"
)

func main() {
	// Parse command-line arguments
	configPath := flag.String("config", "./config.json", "path to the config file")
	flag.Parse()

	// Load configuration or use default if the file is not found
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.ErrorLogger.Printf("Server side: Failed to load configuration: %v", err)
		logger.InfoLogger.Println("Using default configuration")

		cfg = &config.Config{
			ServerPort: "6379", // Default port number
			LogLevel:   "info", // Default log level
		}
	}

	// Initialize and start the server with configuration settings
	srv := server.NewServer(cfg.ServerPort)
	fmt.Printf("Starting server on port %s...\n", cfg.ServerPort)
	if err := srv.Start(); err != nil {
		logger.ErrorLogger.Printf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
