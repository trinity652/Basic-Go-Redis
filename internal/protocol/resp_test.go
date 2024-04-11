package protocol

import (
	"bufio"
	"strings"
	"testing"
)

func TestSerialize(t *testing.T) {
	command := "SET"
	args := []string{"key", "value"}
	expected := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"

	result := Serialize(command, args)
	if result != expected {
		t.Errorf("Serialize(%s, %v) = %s, want %s", command, args, result, expected)
	}
}

func TestDeserialize(t *testing.T) {
	input := "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n"
	reader := bufio.NewReader(strings.NewReader(input))

	command, args, err := Deserialize(reader)
	if err != nil {
		t.Fatalf("Deserialize() error: %v", err)
	}

	if command != "GET" || len(args) != 1 || args[0] != "key" {
		t.Errorf("Deserialize() = %s, %v, want %s, %v", command, args, "GET", []string{"key"})
	}
}
