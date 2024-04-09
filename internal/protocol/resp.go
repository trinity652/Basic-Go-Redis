// File: internal/protocol/resp.go

package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Serialize takes a command and its arguments and returns the RESP serialized string
func Serialize(command string, args []string) string {
	var resp strings.Builder

	// Serialize the command and arguments into the RESP array format
	resp.WriteString(fmt.Sprintf("*%d\r\n", 1+len(args)))                 // Array length
	resp.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(command), command)) // Command itself

	for _, arg := range args {
		resp.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)) // Each argument
	}
	//fmt.Println("Resp Serialization:", resp.String())
	return resp.String()
}

// Deserialize reads from a connection and parses the RESP command
func Deserialize(reader *bufio.Reader) (string, []string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", nil, err
	}
	line = strings.TrimRight(line, "\r\n") // Trim CR and LF

	if !strings.HasPrefix(line, "*") {
		return "", nil, errors.New("deserialization protocol error: expected '*'")
	}

	count, err := strconv.Atoi(line[1:])
	if err != nil {
		return "", nil, fmt.Errorf("deserialization protocol error: invalid array length")
	}

	var command string
	var args []string

	for i := 0; i < count; i++ {
		line, err = reader.ReadString('\n')
		if err != nil {
			return "", nil, err
		}
		line = strings.TrimRight(line, "\r\n") // Trim CR and LF

		if !strings.HasPrefix(line, "$") {
			return "", nil, errors.New("deserialization protocol error: expected '$'")
		}

		length, err := strconv.Atoi(line[1:])
		if err != nil {
			return "", nil, errors.New("deserialization protocol error: invalid bulk string length")
		}

		if length == -1 { // Handle null bulk string
			args = append(args, "")
			continue
		}

		bulk := make([]byte, length)
		_, err = io.ReadFull(reader, bulk)
		if err != nil {
			return "", nil, err
		}

		_, err = reader.Discard(2) // Discard CRLF after bulk string
		if err != nil {
			return "", nil, err
		}

		arg := string(bulk)
		if i == 0 {
			command = arg
		} else {
			args = append(args, arg)
		}
	}

	return command, args, nil
}

func ReadFullResponse(reader *bufio.Reader) (string, error) {
	var response strings.Builder

	for {
		line, err := reader.ReadString('\n')
		//fmt.Println("Line:", line)
		if err != nil {
			if err == io.EOF && response.Len() > 0 {
				break // End of file reached, return what we have
			}
			return "", err
		}

		// Break the loop if the line is empty
		if line == "\r\n" || line == "\n" {
			break
		}

		response.WriteString(line)

		// Check if we've reached the end of a Redis response
		if line[0] == '+' || line[0] == '-' || line[0] == ':' || line[0] == '_' {
			break
		} else if line[0] == '$' { // Bulk string
			length, _ := strconv.Atoi(strings.TrimSpace(line[1:]))
			if length == -1 {
				break // Handle nil bulk string
			}
			// Read the bulk string content
			content := make([]byte, length+2) // +2 for \r\n
			_, err = io.ReadFull(reader, content)
			if err != nil {
				return "", err
			}
			response.Write(content)
			//fmt.Println("Bulk String:", response.String())
		} else if line[0] == '*' { // Array
			count, _ := strconv.Atoi(strings.TrimSpace(line[1:]))
			if count == -1 {
				// Handle nil array
				continue
			}
			for i := 0; i < count; i++ {
				element, err := reader.ReadString('\n')
				if err != nil {
					return "", err
				}
				response.WriteString(element)

			}
		}

	}

	return response.String(), nil
}

func ConvertRESPToReadable(response string) string {

	response = strings.TrimSuffix(response, "\r\n") // Remove trailing CRLF
	var new_response strings.Builder
	if strings.HasPrefix(response, "+") {
		// Simple string
		return strings.TrimPrefix(response, "+")
	} else if strings.HasPrefix(response, ":") {
		// Integer
		return strings.TrimPrefix(response, ":")
	} else if strings.HasPrefix(response, "$") {
		// Bulk string
		parts := strings.SplitN(response, "\r\n", 2)

		if len(parts) > 1 {
			return parts[1]
		}
	} else if strings.HasPrefix(response, "*") {
		// Array
		parts := strings.SplitN(response, "\r\n", 2)
		if len(parts) > 1 {
			arrayBody := parts[1]
			arrayElements := strings.Split(arrayBody, "\r\n")
			k := 0
			for _, element := range arrayElements {
				if element != "" && !strings.HasPrefix(element, "$") {
					k++
					formattedString := fmt.Sprintf("%d) \"%s\"\n", k, element)
					new_response.WriteString(formattedString)

				}

			}
		}
		return new_response.String()
	}
	return response
}
