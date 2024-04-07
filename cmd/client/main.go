// File: cmd/client/main.go

package main

import (
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

		_, err = conn.Write([]byte(trimmedInput + "\r\n"))
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
