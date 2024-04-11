package server

import (
	"basic-go-redis/internal/protocol"
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// Helper function to start the server for testing
func startTestServer(port string) *Server {
	server := NewServer(port)
	go func() {
		if err := server.Start(); err != nil {
			panic(fmt.Sprintf("Failed to start server: %v", err))
		}
	}()
	time.Sleep(time.Second) // Give the server time to start
	return server
}

// Helper function to send a command to the server and receive a response
func sendCommand(conn net.Conn, command string) (string, error) {
	_, err := conn.Write([]byte(command))
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(conn)
	response, err := protocol.ReadFullResponse(reader)
	return response, err
}

func TestServer_SET_GET(t *testing.T) {
	port := "12345"
	server := startTestServer(port)
	defer server.Close()

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		t.Fatalf("Failed to connect to server on port %s: %v", port, err)
	}
	defer conn.Close()

	// Test SET command
	setCommand := "*3\r\n$3\r\nSET\r\n$7\r\ntestkey\r\n$9\r\ntestvalue\r\n"
	setResponse, err := sendCommand(conn, setCommand)
	if err != nil || setResponse != "+OK\r\n" {
		t.Errorf("SET command failed: %v, response: %s", err, setResponse)
	}

	// Test GET command
	getCommand := "*2\r\n$3\r\nGET\r\n$7\r\ntestkey\r\n"
	getResponse, err := sendCommand(conn, getCommand)
	if err != nil || !strings.Contains(getResponse, "$9\r\ntestvalue\r\n") {
		t.Errorf("GET command failed: %v, response: %s", err, getResponse)
	}
}

func TestServer_DEL(t *testing.T) {
	port := "12345"
	server := startTestServer(port)
	defer server.Close()

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		t.Fatalf("Failed to connect to server on port %s: %v", port, err)
	}
	defer conn.Close()

	// First, set a key to delete
	sendCommand(conn, "*3\r\n$3\r\nSET\r\n$6\r\ndelete\r\n$5\r\nvalue\r\n")

	// Now, try to delete the key
	delResponse, err := sendCommand(conn, "*2\r\n$3\r\nDEL\r\n$6\r\ndelete\r\n")
	if err != nil || delResponse != ":1\r\n" {
		t.Errorf("DEL command failed: %v, response: %s", err, delResponse)
	}
}

func TestServer_KEYS(t *testing.T) {
	port := "12345"
	server := startTestServer(port)
	defer server.Close()

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		t.Fatalf("Failed to connect to server on port %s: %v", port, err)
	}
	defer conn.Close()

	// Set a few keys
	sendCommand(conn, "*3\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$5\r\nvalue\r\n")
	sendCommand(conn, "*3\r\n$3\r\nSET\r\n$4\r\nkey2\r\n$5\r\nvalue\r\n")

	// Get keys list
	keysResponse, err := sendCommand(conn, "*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n")
	if err != nil || !strings.Contains(keysResponse, "key1") || !strings.Contains(keysResponse, "key2") {
		t.Errorf("KEYS command failed: %v, response: %s", err, keysResponse)
	}
}

func TestServer_EXPIRE_TTL(t *testing.T) {
	port := "12345"
	server := startTestServer(port)
	defer server.Close()

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		t.Fatalf("Failed to connect to server on port %s: %v", port, err)
	}
	defer conn.Close()

	// Set a key to test expiration
	sendCommand(conn, "*3\r\n$3\r\nSET\r\n$7\r\ntempkey\r\n$5\r\nvalue\r\n")

	// Set expiration
	expireResponse, err := sendCommand(conn, "*3\r\n$6\r\nEXPIRE\r\n$7\r\ntempkey\r\n$1\r\n2\r\n") // 2 second
	if err != nil || expireResponse != ":1\r\n" {
		t.Errorf("EXPIRE command failed: %v, response: %s", err, expireResponse)
	}

	// Check TTL
	ttlResponse, err := sendCommand(conn, "*2\r\n$3\r\nTTL\r\n$7\r\ntempkey\r\n")
	if err != nil {
		t.Errorf("TTL command failed: %v, response: %s", err, ttlResponse)
	}
}

func TestServer_ZADD_ZRANGE(t *testing.T) {
	port := "12345"
	server := startTestServer(port)
	defer server.Close()

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%s", port))
	if err != nil {
		t.Fatalf("Failed to connect to server on port %s: %v", port, err)
	}
	defer conn.Close()

	// Add members to a sorted set
	sendCommand(conn, "*4\r\n$4\r\nZADD\r\n$6\r\nmyzset\r\n$1\r\n1\r\n$7\r\nmember1\r\n")
	sendCommand(conn, "*4\r\n$4\r\nZADD\r\n$6\r\nmyzset\r\n$1\r\n2\r\n$7\r\nmember2\r\n")

	// Get range of members
	zrangeResponse, err := sendCommand(conn, "*4\r\n$6\r\nZRANGE\r\n$6\r\nmyzset\r\n$1\r\n0\r\n$1\r\n1\r\n")
	if err != nil || !strings.Contains(zrangeResponse, "member1") || !strings.Contains(zrangeResponse, "member2") {
		t.Errorf("ZRANGE command failed: %v, response: %s", err, zrangeResponse)
	}
}
