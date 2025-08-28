# MC-MOTD

A lightweight Minecraft placeholder server that provides dynamic server status responses and webhook notifications for connection attempts.

## Features

- **Dynamic MOTD (Message of the Day)**: Automatically changes server status messages based on connection state
- **Smart Status Management**: Shows different messages when server is sleeping vs. starting up
- **Webhook Notifications**: Optional webhook support for connection attempt notifications
- **Configurable Server Info**: Customize version, protocol, max players, and status messages
- **Lightweight**: Built with Go for minimal resource usage
- **Docker Support**: Ready-to-use Docker containers

## How It Works

MC-MOTD acts as a placeholder Minecraft server that responds to server list pings and login attempts. It's designed to:

1. **Show Server Status**: Responds to Minecraft client server list requests with customizable MOTDs
2. **Track Connection Attempts**: Detects when players try to connect and can notify external systems
3. **State Management**: Transitions between "sleeping" and "starting" states based on activity
4. **Webhook Integration**: Sends HTTP notifications when connection attempts occur

**Note**: This is not a proxy or router - it's a standalone placeholder server that doesn't connect to or forward traffic to other Minecraft servers.

## Installation

### Using Go

```bash
go install github.com/wroud/mc-motd/cmd/mc-motd@latest
```

### Using Docker

```bash
docker run -p 25565:25565 your-registry/mc-motd
```

### Building from Source

```bash
git clone https://github.com/wroud/mc-motd.git
cd mc-motd
go build ./cmd/mc-motd
```

## Configuration

MC-MOTD can be configured using command-line flags or environment variables.

### Command Line Options

```bash
./mc-motd --help
```

### Key Configuration Options

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `--port` | `PORT` | `25565` | Port to listen for Minecraft connections |
| `--sleeping-motd` | `SLEEPING_MOTD` | `🌙 Server sleeping, join to wake up!` | MOTD when server is sleeping |
| `--starting-motd` | `STARTING_MOTD` | `⚡ Server starting up...` | MOTD when server is starting |
| `--starting-timeout` | `STARTING_TIMEOUT` | `300` | Seconds to show starting MOTD (5 minutes) |
| `--max-players` | `MAX_PLAYERS` | `20` | Max players shown in server list |
| `--version` | `VERSION` | `1.21.8` | Minecraft version displayed |
| `--protocol` | `PROTOCOL` | `772` | Protocol version number |
| `--webhook-url` | `WEBHOOK_URL` | | HTTP endpoint for connection attempt notifications |
| `--webhook-require-user` | `WEBHOOK_REQUIRE_USER` | `false` | Only send webhook for actual user connections |

### Example Usage

```bash
# Basic placeholder server
./mc-motd

# Custom MOTDs for different states
./mc-motd \
  --sleeping-motd "💤 Server is sleeping - join to wake it up!" \
  --starting-motd "🚀 Booting up the server..." \
  --starting-timeout 180

# With webhook notifications for connection attempts
./mc-motd \
  --webhook-url "https://your-webhook-endpoint.com/notify" \
  --webhook-require-user true
```

## Docker Usage

### Using the Dockerfile

```bash
# Build the image
docker build -t mc-motd .

# Run with default settings
docker run -p 25565:25565 mc-motd

# Run with custom configuration
docker run -p 25565:25565 \
  -e SLEEPING_MOTD="💤 Server is hibernating..." \
  -e STARTING_MOTD="🔥 Firing up the server!" \
  -e STARTING_TIMEOUT=120 \
  mc-motd
```

### Docker Compose Example

```yaml
version: '3.8'
services:
  mc-motd:
    build: .
    ports:
      - "25565:25565"
    environment:
      - SLEEPING_MOTD=🌙 Server sleeping, join to wake up!
      - STARTING_MOTD=⚡ Server starting up...
      - STARTING_TIMEOUT=300
      - MAX_PLAYERS=20
      - VERSION=1.21.8
      - PROTOCOL=772
    restart: unless-stopped
```

## Webhook Integration

When configured with a webhook URL, MC-MOTD will send POST requests with connection attempt information:

```json
{
  "event": "connection_attempt",
  "timestamp": "2025-08-29T10:30:00Z",
  "client_ip": "192.168.1.100",
  "has_user": true
}
```

### Webhook Configuration

- Set `--webhook-url` to your HTTP endpoint
- Use `--webhook-require-user true` to only receive notifications for actual user connections (not server list pings)

## Protocol Support

Currently supports Minecraft protocol version 772 (1.21.8) by default. You can configure different versions using the `--version` and `--protocol` flags.

Common protocol versions:
- 772: Minecraft 1.21.8
- 770: Minecraft 1.21.5
- 769: Minecraft 1.21.4

## Development

### Prerequisites

- Go 1.24.4 or later

### Building

```bash
# Download dependencies
go mod download

# Run tests
make test

# Build binary
go build ./cmd/mc-motd

# Build for release
make release
```

### Project Structure

```
├── cmd/mc-motd/          # Main application entry point
├── server/               # Core server logic
│   ├── configs.go        # Configuration structures
│   ├── server.go         # Main server implementation
│   ├── connector.go      # Connection handling
│   ├── motd_manager.go   # MOTD state management
│   ├── notifier.go       # Notification interfaces
│   └── webhook_notifier.go # Webhook implementation
├── mcproto/              # Minecraft protocol handling
│   ├── decode.go         # Protocol decoding
│   ├── read.go           # Data reading utilities
│   ├── types.go          # Protocol type definitions
│   └── write.go          # Data writing utilities
├── Dockerfile            # Docker build configuration
└── Makefile             # Build automation
```

## Logging

MC-MOTD uses structured logging with different levels:

- `--debug`: Enable debug logs
- `--trace`: Enable trace logs (most verbose)

## Use Cases

This placeholder server is ideal for:

- **Server Wake-up Systems**: Notify external systems when players attempt to connect to start up actual game servers
- **Maintenance Pages**: Show custom messages when your main server is down or updating
- **Load Balancing**: Use webhooks to trigger server provisioning based on connection attempts
- **Analytics**: Track connection patterns and player interest
- **Development/Testing**: Test Minecraft client integrations without running a full server

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Based on [mc-router](https://github.com/itzg/mc-router) - thanks for providing the foundation for this project
- Built with [go-flagsfiller](https://github.com/itzg/go-flagsfiller) for configuration management
- Uses [logrus](https://github.com/sirupsen/logrus) for structured logging