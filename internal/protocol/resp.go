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

	return resp.String()
}

// Deserialize reads from a connection and parses the RESP command
func Deserialize(reader *bufio.Reader) (string, []string, error) {
	// Read the array length
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", nil, err
	}

	if line[0] != '*' {
		return "", nil, errors.New("protocol error: expected '*'")
	}

	count, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		return "", nil, fmt.Errorf("protocol error: invalid array length")
	}

	var command string
	var args []string

	for i := 0; i < count; i++ {
		// Read the bulk string length
		line, err = reader.ReadString('\n')
		if err != nil {
			return "", nil, err
		}

		if line[0] != '$' {
			return "", nil, errors.New("protocol error: expected '$'")
		}

		length, err := strconv.Atoi(strings.TrimSpace(line[1:]))
		if err != nil {
			return "", nil, errors.New("protocol error: invalid bulk string length")
		}

		// Read the bulk string
		bulk := make([]byte, length)
		_, err = io.ReadFull(reader, bulk)
		if err != nil {
			return "", nil, err
		}

		// Discard the CRLF
		if _, err = reader.Discard(2); err != nil {
			return "", nil, err
		}

		if i == 0 {
			command = string(bulk)
		} else {
			args = append(args, string(bulk))
		}
	}

	return command, args, nil
}
