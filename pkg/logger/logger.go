// File: pkg/logger/logger.go

package logger

import (
	"log"
	"os"
)

var (
	// InfoLogger for logging informational messages
	InfoLogger *log.Logger
	// ErrorLogger for logging error messages
	ErrorLogger *log.Logger
)

func init() {
	// Open a file for logging
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}

	InfoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
