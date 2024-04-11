// File: internal/server/server.go

package server

import (
	"basic-go-redis/internal/protocol"
	"basic-go-redis/internal/store"
	"basic-go-redis/pkg/logger"
	"bufio"
	"fmt"
	"net"
	"strconv"
	"sync"
)

type Server struct {
	store        *store.InMemoryStore
	port         string
	connections  map[net.Conn]bool
	connLock     sync.Mutex
	shutdownChan chan struct{}
	listener     net.Listener
	wg           sync.WaitGroup // WaitGroup to wait for goroutines to finish
}

func NewServer(port string) *Server {
	return &Server{
		store:        store.NewInMemoryStore(),
		port:         port,
		connections:  make(map[net.Conn]bool),
		shutdownChan: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}

	fmt.Printf("Server listening on port %s\n", s.port)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Check if we should stop accepting new connections
			select {
			case <-s.shutdownChan:
				fmt.Printf("Server is shutting down")
				return nil // Server is shutting down, exit the loop
			default:
				fmt.Printf("Error accepting connection: %v\n", err)
				continue
			}
		}

		s.connLock.Lock()
		s.connections[conn] = true
		s.connLock.Unlock()

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	s.wg.Add(1)       // Increment the WaitGroup counter
	defer s.wg.Done() // Decrement the counter when the goroutine completes

	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		command, args, err := protocol.Deserialize(reader)
		if err != nil {
			// Log error and exit the goroutine
			logger.ErrorLogger.Printf("Error deserializing command: %v\n", err)
			return
		}

		response := s.executeCommand(command, args)
		if _, err := conn.Write([]byte(response)); err != nil {
			// Log error and exit the goroutine
			logger.ErrorLogger.Printf("Error in sending response: %v\n", err)
			return
		}
	}
}

func (s *Server) Close() error {
	// Signal the listener to stop accepting new connections
	close(s.shutdownChan)

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return err
		}
	}

	s.connLock.Lock()
	for conn := range s.connections {
		conn.Close() // Ignore error
		delete(s.connections, conn)
	}
	s.connLock.Unlock()

	// Wait for all handleConnection goroutines to finish
	s.wg.Wait()

	return nil
}

func (s *Server) executeCommand(command string, args []string) string {
	switch command {
	case "SET":
		if len(args) < 2 {
			return "-ERR wrong number of arguments for 'SET' command\r\n"
		}

		key := args[0]
		value := args[1]
		flags := args[2:] // All remaining arguments are considered as flags or TTL

		// Call the Set function with flags
		response := s.store.Set(key, value, flags...)
		if response == "+0\r\n" {
			return "-ERR condition not met for 'SET' command\r\n"
		}
		return response

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
		return fmt.Sprintf(":%d\r\n\r\n", deleted)

	case "KEYS":
		if len(args) != 1 {
			return "-ERR wrong number of arguments for 'KEYS' command\r\n"
		}
		keys := s.store.Keys(args[0])
		response := fmt.Sprintf("*%d\r\n", len(keys))
		for _, key := range keys {
			response += fmt.Sprintf("$%d\r\n%s\r\n", len(key), key)
		}
		response += "\r\n"
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
		return fmt.Sprintf(":%d\r\n\r\n", result)

	case "TTL":
		if len(args) != 1 {
			return "-ERR wrong number of arguments for 'TTL' command\r\n"
		}
		ttl := s.store.TTL(args[0])
		return fmt.Sprintf(":%d\r\n\r\n", ttl)

	case "ZADD":
		if len(args) < 3 {
			return "-ERR wrong number of arguments\r\n"
		}
		score, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return "-ERR invalid score\r\n"
		}
		added := s.store.ZAdd(args[0], score, args[2])
		return fmt.Sprintf(":%d\r\n\r\n", added)

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

			return "-ERR invalid stop argument" + "\r\n"
		}
		members := s.store.ZRange(args[0], start, stop)
		return fmt.Sprintf("+%v\r\n\r\n", members)

	default:
		return "-ERR unknown command\r\n"
	}
}
