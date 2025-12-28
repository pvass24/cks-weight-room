# CKS Weight Room

Kubernetes Security Practice Environment for CKS certification preparation.

## Installation

### macOS (Homebrew)

```bash
brew install cks-weight-room
```

### Linux / macOS (curl)

```bash
curl -fsSL https://install.cks-weight-room.com | bash
```

### Manual Installation

Download the appropriate binary for your platform from [GitHub Releases](https://github.com/patrickvassell/cks-weight-room/releases):

- **macOS (Intel)**: `cks-weight-room-darwin-amd64`
- **macOS (Apple Silicon)**: `cks-weight-room-darwin-arm64`
- **Linux (x86_64)**: `cks-weight-room-linux-amd64`
- **Linux (ARM64)**: `cks-weight-room-linux-arm64`

Make the binary executable and move it to your PATH:

```bash
chmod +x cks-weight-room-*
sudo mv cks-weight-room-* /usr/local/bin/cks-weight-room
```

## Usage

Start CKS Weight Room:

```bash
cks-weight-room
```

The application will start a web server on `http://127.0.0.1:3000`.

### Command Line Options

- `--version`: Display version information
- `--port <port>`: Specify server port (default: 3000)

## Requirements

- Docker Desktop (for Kubernetes cluster provisioning)
- macOS or Linux

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- npm

### Building from Source

```bash
# Build for all platforms
./scripts/build.sh

# Build for current platform only
go build -ldflags "-X main.version=$(git describe --tags)" -o cks-weight-room .
```

### Running Tests

```bash
# Go tests
go test ./...

# Frontend tests
cd web && npm test
```

## Architecture

CKS Weight Room is a single binary application that embeds a Next.js frontend:

- **Backend**: Go 1.21+ with standard library HTTP server
- **Frontend**: Next.js 14 with static export
- **Database**: SQLite (local progress tracking)
- **Kubernetes**: KIND (Kubernetes in Docker)

## License

MIT

## Support

For issues and feature requests, please use [GitHub Issues](https://github.com/patrickvassell/cks-weight-room/issues).
