// File: cmd/client/main.go

package main

import (
	"basic-go-redis/internal/protocol"
	"basic-go-redis/pkg/config"
	"basic-go-redis/pkg/logger"
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

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
		//fmt.Println("Serialized input:", serializedInput)

		_, err = conn.Write([]byte(serializedInput))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending command to server: %v\n", err)
			continue
		}

		response, err := protocol.ReadFullResponse(bufio.NewReader(conn))
		//fmt.Println("Response received by client:", response)
		logger.InfoLogger.Printf("Response received by client: %s", response)
		if err != nil {
			logger.ErrorLogger.Printf("Error in recieving response: %v\n", err)
			continue
		}
		readable := protocol.ConvertRESPToReadable(response)
		fmt.Println(readable)
	}
}
