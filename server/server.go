package main

import (
    "bufio"
    "fmt"
    "net"
    "strings"
)

type InMemoryStore struct {
    data map[string]string
}

func NewInMemoryStore() *InMemoryStore {
    return &InMemoryStore{
        data: make(map[string]string),
    }
}

func (store *InMemoryStore) executeCommand(command string, args []string) string {
    switch command {
    case "SET":
        if len(args) != 2 {
            return "-ERR wrong number of arguments for 'set' command\r\n"
        }
        store.data[args[0]] = args[1]
        return "+OK\r\n"
    case "GET":
        if len(args) != 1 {
            return "-ERR wrong number of arguments for 'get' command\r\n"
        }
        value, ok := store.data[args[0]]
        if !ok {
            return "$-1\r\n"
        }
        return fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
    // Add other commands here
    default:
        return "-ERR unknown command\r\n"
    }
}

func handleConnection(conn net.Conn, store *InMemoryStore) {
    defer conn.Close()

    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        input := scanner.Text()
        parts := strings.Split(input, " ")
        command := strings.ToUpper(parts[0])
        args := parts[1:]

        response := store.executeCommand(command, args)
        conn.Write([]byte(response))
    }
}

func main() {
    store := NewInMemoryStore()
    listener, err := net.Listen("tcp", ":6379")
    if err != nil {
        fmt.Println("Error starting server:", err)
        return
    }
    defer listener.Close()

    fmt.Println("Server is running on port 6379")
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            continue
        }
        go handleConnection(conn, store)
    }
}

