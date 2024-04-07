// File: cmd/client/main.go

package main

import (
	"basic-go-redis/internal/protocol" // Adjust the import path as per your module name
	"basic-go-redis/pkg/config"
	"basic-go-redis/pkg/logger"
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

// displayResponse processes the server response and formats it for display just like redis-cli
func displayResponse(response string) {
	response = strings.TrimSuffix(response, "\r\n") // Remove trailing CRLF

	// Strip protocol-specific characters for simple strings
	if strings.HasPrefix(response, "+") {
		fmt.Println(strings.TrimPrefix(response, "+"))
	} else if strings.HasPrefix(response, ":") {
		// For integer responses, remove the ":" prefix
		fmt.Println(strings.TrimPrefix(response, ":"))
	} else if strings.HasPrefix(response, "$") {
		fmt.Println(strings.TrimPrefix(response, "$"))
	} else if strings.HasPrefix(response, "-") {
		// For errors, remove the "-" prefix
		fmt.Println(strings.TrimPrefix(response, "-"))
	} else if strings.HasPrefix(response, "*") {
		// For arrays, handle each element accordingly
		// This is a simplified example; you'll need to parse the array properly
		fmt.Println("Array response:", response)
	} else {
		// If it doesn't match any known RESP type, print as is
		fmt.Println(response)
	}
}

func main() {

	configPath := flag.String("config", "./config.json", "path to the config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath) // Specify the path to your configuration file
	if err != nil {
		logger.ErrorLogger.Printf("Client side: Failed to load configuration: %v", err)
		logger.InfoLogger.Println("Using default configuration")

		cfg = &config.Config{
			ServerHost: "localhost", // Default server host
			ServerPort: "6379",      // Default port number
			LogLevel:   "info",      // Default log level
		}
	}

	serverAddress := cfg.ServerHost + ":" + cfg.ServerPort

	conn, err := net.Dial("tcp", serverAddress) // This could be configured through arguments or environment variables

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	consoleReader := bufio.NewReader(os.Stdin)
	fmt.Printf("Connected to Go-Redis server at %s. Type 'exit' to quit.\n", serverAddress)

	for {
		fmt.Print("go-redis-cli> ")
		userInput, err := consoleReader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			continue
		}

		trimmedInput := strings.TrimSpace(userInput)
		if trimmedInput == "exit" {
			break
		}

		parts := strings.Split(trimmedInput, " ")
		command := parts[0]
		args := parts[1:]

		serializedInput := protocol.Serialize(command, args)

		_, err = conn.Write([]byte(serializedInput + "\r\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending command to server: %v\n", err)
			continue
		}

		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response from server: %v\n", err)
			continue
		}

		displayResponse(response)
	}
}
