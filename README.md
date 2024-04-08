
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

With the client running, you can enter Redis commands, such as:

```
go-redis-cli> SET mykey myvalue
OK
go-redis-cli> GET mykey
myvalue
```

Type `exit` to quit the client.

## Development

- The server's main logic is located in `cmd/server/main.go`.
- Client implementation can be found in `cmd/client/main.go`.
- RESP protocol handling is in `internal/protocol/resp.go`.
- The in-memory store logic is in `internal/store/store.go`.
- Configuration handling is managed in `pkg/config/config.go`.
- Logging utilities are located in `pkg/logger/logger.go`.
