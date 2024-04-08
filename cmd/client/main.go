// File: cmd/client/main.go

package main

import (
	"basic-go-redis/internal/protocol" // Adjust the import path as per your module name
	"basic-go-redis/pkg/config"
	"basic-go-redis/pkg/logger"
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func readFullResponse(reader *bufio.Reader) (string, error) {
	var response strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return response.String(), err
		}
		response.WriteString(line)

		// Check the first character to determine the type of the response
		if line[0] == '+' || line[0] == '-' || line[0] == ':' || strings.HasPrefix(line, "$-1") {
			// For simple strings, errors, integers, or empty bulk strings, break after processing
			break
		} else if line[0] == '$' {
			// Process bulk strings
			length, _ := strconv.Atoi(strings.TrimSpace(line[1:]))
			if length > 0 {
				content := make([]byte, length)
				_, err := io.ReadFull(reader, content)
				if err != nil {
					return response.String(), err
				}
				response.Write(content)
				// Read and discard the trailing CRLF after the bulk string
				_, err = reader.Discard(2)
				if err != nil {
					return response.String(), err
				}
			}
			return response.String(), nil
		} else if line[0] == '*' {
			// Process arrays
			count, _ := strconv.Atoi(strings.TrimSpace(line[1:]))
			if count <= 0 {
				break
			}
			// More logic here to handle the array elements, potentially requiring a recursive approach
		} else {
			// Unexpected response type
			return response.String(), fmt.Errorf("unexpected response type: %s", line)
		}
	}

	return response.String(), nil

}

// displayResponse processes the server response and formats it for display just like redis-cli
func displayResponse(response string) {
	response = strings.TrimSuffix(response, "\r\n") // Remove trailing CRLF
	if strings.HasPrefix(response, "+") {
		// Simple string
		fmt.Println(strings.TrimPrefix(response, "+"))
	} else if strings.HasPrefix(response, ":") {
		// Integer
		fmt.Println(strings.TrimPrefix(response, ":"))
	} else if strings.HasPrefix(response, "$") {
		// Bulk string
		parts := strings.SplitN(response, "\r\n", 2)

		if len(parts) > 1 {
			fmt.Println(parts[1])
		}
	} else if strings.HasPrefix(response, "*") {
		// Array
		parts := strings.SplitN(response, "\r\n", 2)
		if len(parts) > 1 {
			arrayBody := parts[1]
			arrayElements := strings.Split(arrayBody, "\r\n")
			for i, element := range arrayElements {
				if element != "" && !strings.HasPrefix(element, "$") {
					fmt.Printf("%d) \"%s\"\n", i, element)
				}
			}
		}
	} else {
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
		//fmt.Println("Serialized input:", serializedInput)

		_, err = conn.Write([]byte(serializedInput))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending command to server: %v\n", err)
			continue
		}

		response, err := readFullResponse(bufio.NewReader(conn))

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response from server: %v\n", err)
			continue
		}
		displayResponse(response)
	}
}
