// File: cmd/client/main.go

package main

import (
	"basic-go-redis/internal/protocol" // Adjust the import path as per your module name
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	serverAddress := "localhost:6379" // This could be configured through arguments or environment variables

	conn, err := net.Dial("tcp", serverAddress)
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

		// Splitting the input into command and arguments
		parts := strings.Split(trimmedInput, " ")
		command := parts[0]
		args := parts[1:]

		// Serialize the input using the RESP protocol
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

		fmt.Print("Response: ", response)
	}
}
