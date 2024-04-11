
# Basic-Go-Redis

Basic-Go-Redis is a simplified Redis-like in-memory database server implemented in Go, supporting a subset of the Redis protocol. This project also includes a client that communicates with the server using the Redis Serialization Protocol (RESP).

## Features

- In-memory key-value storage
- Support for commands: `SET`, `GET`, `DEL`, `EXPIRE`, `KEYS`, `TTL`, `ZADD`, `ZRANGE`
- RESP protocol for client-server communication
- Configurable server settings

### Prerequisites

- Go (version 1.14 or later)

### Installation

Clone the repository:

```bash
git clone https://github.com/yourusername/go-redis.git
cd basic-go-redis
```

Build the server and client:

```bash
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
```

You could also do the following:  

To run the server:  
```bash
go run cmd/server/main.go

```
To run the client:  
```bash
go run cmd/client/main.go

```

### Configuration

Edit `config.json` to specify the server port and logging level:

```json
{
    "server_port": "6379",
    "log_level": "info"
}
```

If `config.json` is not present, the server defaults to port `6379` and log level `info`.

### Running the Server

Navigate to the `bin` directory and run:

```bash
./server
```

### Running the Client

In a separate terminal window, navigate to the `bin` directory and run:

```bash
./client
```

## Usage

With the client running, you can enter Redis commands, such as(note that it supports only completely uppercase commands):

```
go-redis-cli> SET keys Abhilasha
OK
go-redis-cli> GET keys
Abhilasha
go-redis-cli> SET key Abhilasha
OK
go-redis-cli> KEYS *
1) "key"
2) "keys"

go-redis-cli> DEL keys
1
go-redis-cli> TTL keys
-1
go-redis-cli> EXPIRE keys 80
1
go-redis-cli> ZADD Myset 1 GulabJamun
1
go-redis-cli> ZRANGE Myset 0 1
[GulabJamun]

```

Type `exit` to quit the client.

## Testing 




## Development

- The server's execution logic is located in `cmd/server/main.go`.
- The in-memory store logic for the server is in `internal/store/store.go`.
- The creation and request handling for the server is in `internal/server/server.go`.
- Client implementation can be found in `cmd/client/main.go`.
- RESP protocol handling is in `internal/protocol/resp.go`.
- Configuration handling is managed in `pkg/config/config.go`.
- Logging utilities are located in `pkg/logger/logger.go`.
