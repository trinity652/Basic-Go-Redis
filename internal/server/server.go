// File: internal/server/server.go

package server

import (
	"basic-go-redis/internal/protocol"
	"basic-go-redis/internal/store"
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
)

type Server struct {
	store *store.InMemoryStore
	port  string
}

func NewServer(port string) *Server {
	return &Server{
		store: store.NewInMemoryStore(), // Initialize the InMemoryStore
		port:  port,
	}
}

func (s *Server) Start() error {

	// Start listening on the specified port
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Server listening on port %s\n", s.port)

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go s.handleConnection(conn) // Go helps to run the handleConnection in a separate goroutine
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		command, args, err := protocol.Deserialize(reader)
		if err != nil {
			log.Printf("Error deserializing command: %v\n", err)
			fmt.Fprintf(conn, "-ERR %s\r\n", err)
			continue
		}

		log.Printf("Received command: %s with args: %v\n", command, args)

		response := s.executeCommand(command, args)
		log.Printf("Received response: %s", response)
		_, err = conn.Write([]byte(response))
		fmt.Println("Response sent")
		if err != nil {
			log.Printf("Error sending response: %v\n", err)
			return
		}
	}
}

func (s *Server) executeCommand(command string, args []string) string {
	switch command {
	case "SET":
		if len(args) < 2 {
			return "-ERR wrong number of arguments for 'SET' command\r\n"
		}
		ttl := 0
		if len(args) > 2 {
			var err error
			ttl, err = strconv.Atoi(args[2])
			if err != nil {
				return "-ERR invalid TTL value\r\n"
			}
		}
		return s.store.Set(args[0], args[1], ttl)

	case "GET":
		if len(args) != 1 {
			return "-ERR wrong number of arguments for 'GET' command\r\n"
		}
		return s.store.Get(args[0])

	case "DEL":
		if len(args) < 1 {
			return "-ERR wrong number of arguments for 'DEL' command\r\n"
		}
		deleted := s.store.Del(args)
		return fmt.Sprintf(":%d\r\n", deleted)

	case "KEYS":
		if len(args) != 1 {
			return "-ERR wrong number of arguments for 'KEYS' command\r\n"
		}
		keys := s.store.Keys(args[0])
		response := fmt.Sprintf("*%d\r\n", len(keys))
		for _, key := range keys {
			response += fmt.Sprintf("$%d\r\n%s\r\n", len(key), key)
		}
		return response

	case "EXPIRE":
		if len(args) != 2 {
			return "-ERR wrong number of arguments for 'EXPIRE' command\r\n"
		}
		seconds, err := strconv.Atoi(args[1])
		if err != nil {
			return "-ERR invalid integer\r\n"
		}
		result := s.store.Expire(args[0], seconds)
		return fmt.Sprintf(":%d\r\n", result)

	case "TTL":
		if len(args) != 1 {
			return "-ERR wrong number of arguments for 'TTL' command\r\n"
		}
		ttl := s.store.TTL(args[0])
		return fmt.Sprintf(":%d\r\n", ttl)

	case "ZADD":
		if len(args) < 3 {
			return "-ERR wrong number of arguments\r\n"
		}
		score, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return "-ERR invalid score\r\n"
		}
		added := s.store.ZAdd(args[0], score, args[2])
		return fmt.Sprintf(":%d\r\n", added)

	case "ZRANGE":
		if len(args) != 3 {
			return "-ERR wrong number of arguments\r\n"
		}
		start, err := strconv.Atoi(args[1])
		if err != nil {
			return "-ERR invalid start argument\r\n"
		}
		stop, err := strconv.Atoi(args[2])
		if err != nil {
			return "-ERR invalid stop argument\r\n"
		}
		members := s.store.ZRange(args[0], start, stop)
		return fmt.Sprintf("+%v\r\n", members)

	default:
		return "-ERR unknown command\r\n"
	}
}
