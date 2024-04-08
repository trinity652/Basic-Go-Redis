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
