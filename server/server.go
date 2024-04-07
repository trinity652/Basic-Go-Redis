package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type InMemoryStore struct {
	data       map[string]string
	sortedSets map[string]map[string]float64
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data:       make(map[string]string),
		sortedSets: make(map[string]map[string]float64),
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
	case "DEL":
		if len(args) < 1 {
			return "-ERR wrong number of arguments for 'del' command\r\n"
		}
		count := 0
		for _, key := range args {
			if _, ok := store.data[key]; ok {
				delete(store.data, key)
				count++
			}
		}
		return fmt.Sprintf(":%d\r\n", count)
	case "EXPIRE":
		if len(args) != 2 {
			return "-ERR wrong number of arguments for 'expire' command\r\n"
		}
		key := args[0]
		seconds, err := strconv.Atoi(args[1])
		if err != nil || seconds < 0 {
			return "-ERR invalid expire time\r\n"
		}
		if _, ok := store.data[key]; !ok {
			return ":0\r\n"
		}
		// Simulate expiration by deleting the key after the specified seconds
		go func() {
			<-time.After(time.Duration(seconds) * time.Second)
			delete(store.data, key)
		}()
		return ":1\r\n"
	case "KEYS":
		if len(args) != 1 {
			return "-ERR wrong number of arguments for 'keys' command\r\n"
		}
		pattern := args[0]
		var keys []string
		for key := range store.data {
			if strings.Contains(key, pattern) {
				keys = append(keys, key)
			}
		}
		response := fmt.Sprintf("*%d\r\n", len(keys))
		for _, key := range keys {
			response += fmt.Sprintf("$%d\r\n%s\r\n", len(key), key)
		}
		return response
	case "TTL":
		if len(args) != 1 {
			return "-ERR wrong number of arguments for 'ttl' command\r\n"
		}
		key := args[0]
		if _, ok := store.data[key]; !ok {
			return ":-2\r\n" // Key does not exist
		}
		return ":-1\r\n" // TTL not supported in in-memory store
	case "ZADD":
		if len(args) < 3 || len(args)%2 != 1 {
			return "-ERR wrong number of arguments for 'zadd' command\r\n"
		}
		key := args[0]
		if store.sortedSets[key] == nil {
			store.sortedSets[key] = make(map[string]float64)
		}
		count := 0
		for i := 1; i < len(args); i += 2 {
			score, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				return "-ERR invalid score value\r\n"
			}
			member := args[i+1]
			store.sortedSets[key][member] = score
			count++
		}
		return fmt.Sprintf(":%d\r\n", count)
	case "ZRANGE":
		if len(args) != 3 {
			return "-ERR wrong number of arguments for 'zrange' command\r\n"
		}
		key := args[0]
		start, err := strconv.Atoi(args[1])
		if err != nil || start < 0 {
			return "-ERR invalid start index\r\n"
		}
		stop, err := strconv.Atoi(args[2])
		if err != nil || stop < 0 {
			return "-ERR invalid stop index\r\n"
		}
		sortedSet, ok := store.sortedSets[key]
		if !ok {
			return "*0\r\n"
		}
		var members []string
		for member := range sortedSet {
			members = append(members, member)
		}
		if stop >= len(members) {
			stop = len(members) - 1
		}
		if start > stop {
			return "*0\r\n"
		}
		members = members[start : stop+1]
		response := fmt.Sprintf("*%d\r\n", len(members))
		for _, member := range members {
			response += fmt.Sprintf("$%d\r\n%s\r\n", len(member), member)
		}
		return response
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
