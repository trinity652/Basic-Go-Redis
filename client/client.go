package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func sendCommand(conn net.Conn, command string) {
	_, err := conn.Write([]byte(command + "\r\n"))
	if err != nil {
		fmt.Println("Error sending command:", err)
		return
	}
}

func readResponse(reader *bufio.Reader) string {
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		return ""
	}

	trimmedResponse := strings.TrimSpace(response)

	// Handle multi-line responses
	if strings.HasPrefix(trimmedResponse, "$") {
		expectedBytes, err := strconv.Atoi(trimmedResponse[1:])
		if err != nil {
			fmt.Println("Error parsing expected bytes:", err)
			return ""
		}

		buffer := make([]byte, expectedBytes)
		_, err = io.ReadFull(reader, buffer)
		if err != nil {
			fmt.Println("Error reading multi-line response:", err)
			return ""
		}

		return string(buffer)
	}

	return trimmedResponse
}

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	inputReader := bufio.NewReader(os.Stdin)

	fmt.Println("Connected to Go-redis server. Enter commands (type 'exit' to quit):")

	for {
		fmt.Print("go-redis-cli> ")
		input, _ := inputReader.ReadString('\n')

		// Trim space and check if user wants to exit
		trimmedInput := strings.TrimSpace(input)
		if trimmedInput == "exit" {
			break
		}

		sendCommand(conn, trimmedInput)
		response := readResponse(reader)
		fmt.Println(response)
	}
}
