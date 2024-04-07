// File: internal/server/server.go

package server

import (
	"basic-go-redis/internal/protocol"
	"basic-go-redis/internal/store"
	"bufio"
	"fmt"
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
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Server listening on port %s\n", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		command, args, err := protocol.Deserialize(reader)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %s\r\n", err)
			continue
		}

		response := s.executeCommand(command, args)
		conn.Write([]byte(response))
	}
}

func (s *Server) executeCommand(command string, args []string) string {
	switch command {
	case "SET":
		if len(args) != 2 {
			return "-ERR wrong number of arguments\r\n"
		}
		return s.store.Set(args[0], args[1])

	case "GET":
		if len(args) != 1 {
			return "-ERR wrong number of arguments\r\n"
		}
		return s.store.Get(args[0])

	case "DEL":
		if len(args) < 1 {
			return "-ERR wrong number of arguments\r\n"
		}
		deleted := s.store.Del(args)
		return fmt.Sprintf(":%d\r\n", deleted)

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
