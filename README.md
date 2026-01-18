# Hytale Server Manager (HSM)

A CLI tool and HTTP service for managing Hytale dedicated server downloads and game sessions. HSM supports both standalone operation for individual users and hosted mode for game server providers (GSPs).

## Features

- **OAuth2 Device Flow Authentication** - Secure login via Hytale's authentication system
- **Server Download Management** - Download and extract Hytale server files
- **Game Session Management** - Create, refresh, and manage game sessions
- **Multi-User Support** - JWT-based authentication for hosting providers
- **Docker Support** - Production-ready container with health checks
- **REST API** - Both plain text and JSON API endpoints

## Installation

### From Source

```bash
git clone https://github.com/your-org/hsm.git
cd hsm
go build -o hsm main.go
```

### Using Docker

```bash
docker build -t hsm .
```

## Usage

### CLI Commands

#### Login

Authenticate with Hytale using OAuth2 device flow:

```bash
# Save session to default location (~/.config/hsm/session.json)
hsm login

# Output session to stdout (useful for automation)
hsm login --stdout
```

#### Download Server Files

Download and extract Hytale server files:

```bash
# Download release version to current directory
hsm download

# Download to specific directory
hsm download --output /path/to/server

# Download prerelease version
hsm download --patchline prerelease
```

#### Start HTTP Server

Run HSM as an HTTP service:

```bash
# Start server on default port (8080)
hsm serve

# Start on custom port
hsm serve --port 3000

# Enable multi-user mode with JWT authentication
hsm serve --jwks-endpoint https://auth.example.com/.well-known/jwks.json
```

### Global Flags

```bash
--session-location string   Path to session.json file (default: ~/.config/hsm/session.json)
```

## Deployment Modes

### Standalone Mode

For individual users running their own Hytale server. Authentication is handled locally via the device flow.

```bash
# 1. Login (opens browser for authentication)
hsm login

# 2. Download server files
hsm download --output ./hytale-server

# 3. Get game session tokens
hsm serve &
curl -X POST http://localhost:8080/game-session
```

### Hosted Mode (Game Server Providers)

For hosting providers managing multiple user servers. Requires a JWKS endpoint for JWT token validation.

When `--jwks-endpoint` is configured:
- All protected endpoints require a valid JWT Bearer token
- Each user (identified by JWT `sub` claim) gets their own isolated session
- Sessions are automatically managed per-user

```bash
hsm serve --jwks-endpoint https://auth.example.com/.well-known/jwks.json
```

## API Reference

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check (returns version info) |

### Protected Endpoints

These endpoints require JWT authentication when running in hosted mode.

#### Plain Text Endpoints (for shell scripts)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/download?patchline=release` | Get server download URL |
| GET | `/version?patchline=release` | Get latest server version |
| POST | `/game-session` | Create game session (returns env format) |

#### JSON API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/session` | Get current session (multi-user only) |
| POST | `/api/v1/session` | Create new game session |
| POST | `/api/v1/session/refresh` | Refresh existing session |
| DELETE | `/api/v1/session` | Delete session |
| GET | `/api/v1/download/version` | Get latest version (JSON) |
| GET | `/api/v1/download/url` | Get download URL (JSON) |

### Response Formats

**Game Session (POST /game-session)**
```bash
HYTALE_SERVER_SESSION_TOKEN="..."
HYTALE_SERVER_IDENTITY_TOKEN="..."
```

**Game Session (POST /api/v1/session)**
```json
{
  "session_token": "...",
  "identity_token": "..."
}
```

## Docker Support

HSM includes full Docker support with a multi-stage build for minimal image size.

### Building the Image

```bash
docker build -t hsm .
```

### Running with Docker

```bash
# Interactive login
docker run -it --rm \
  -v $(pwd)/config:/home/hsm/.config \
  hsm login

# Download server files
docker run -it --rm \
  -v $(pwd):/data \
  -v $(pwd)/config:/home/hsm/.config \
  hsm download

# Run as HTTP server
docker run -d \
  -p 8080:8080 \
  -v $(pwd):/data \
  -v $(pwd)/config:/home/hsm/.config \
  hsm serve
```

### Docker Compose

```yaml
services:
  hsm:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - ./config:/home/hsm/.config
    command: serve --jwks-endpoint=http://auth-service/jwks.json
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HSM_PORT` | Server port (used in health check) | `8080` |

## Examples

### Example 1: Standalone Setup

For personal use or development:

```bash
#!/bin/bash
# standalone-setup.sh

# Build or install HSM
docker build -t hsm .

HSM="docker run -it --rm -v $PWD:/data -v $PWD/config:/home/hsm/.config hsm"

# Login (opens browser)
$HSM login

# Download server files
$HSM download
```

### Example 2: Hosted Mode with JWT Authentication

For game server providers with existing authentication:

```bash
#!/bin/bash
# Setup: Generate JWKS for testing (using step-cli)
docker run --rm -it -v $PWD:/home/step smallstep/step-cli \
  step crypto jwk create jwk.pub.json jwk.json --no-password --insecure

cat jwk.pub.json | docker run --rm -i -v $PWD:/home/step smallstep/step-cli \
  step crypto jwk keyset add jwks.json

# Generate a test JWT token
export JWT_TOKEN=$(docker run --rm -v $PWD:/home/step smallstep/step-cli \
  step crypto jwt sign \
  --key jwk.json \
  --iss "auth.example.com" \
  --aud "api.example.com" \
  --sub gsp-user --subtle)
```

**docker-compose.yml:**
```yaml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "8081:80"
    volumes:
      - .:/usr/share/nginx/html/

  hsm:
    build: .
    depends_on:
      - nginx
    volumes:
      - .:/data
    command: serve --jwks-endpoint=http://nginx/jwks.json
    ports:
      - "8080:8080"
```

### Example 3: GSP Integration Scripts

Install Hytale server via HSM API:

```bash
#!/bin/bash
# gsp-install-hytale-server.sh
set -e

# HSM_URL must be set
if [ -z "$HSM_URL" ]; then
    echo "HSM_URL is not set"
    exit 1
fi

CURL_ARGS=(-sSf)
if [ -n "$JWT_TOKEN" ]; then
    CURL_ARGS+=(-H "Authorization: Bearer $JWT_TOKEN")
fi

# Get download URL from HSM
DOWNLOAD_URL=$(curl "${CURL_ARGS[@]}" "$HSM_URL/download?patchline=${PATCHLINE:-release}")

# Download and extract
curl -sSfL -o hytale-server.zip "$DOWNLOAD_URL"
unzip -o hytale-server.zip
rm hytale-server.zip

echo "Hytale server installed successfully"
```

Start Hytale server with session tokens:

```bash
#!/bin/bash
# gsp-start-hytale-server.sh
set -e

if [ -z "$HSM_URL" ]; then
    echo "HSM_URL is not set"
    exit 1
fi

CURL_ARGS=(-sSf -L -o .env -X POST)
if [ -n "$JWT_TOKEN" ]; then
    CURL_ARGS+=(-H "Authorization: Bearer $JWT_TOKEN")
fi

# Fetch game session and save as .env
curl "${CURL_ARGS[@]}" "$HSM_URL/game-session"
source .env

# Start the server
java -jar Server/HytaleServer.jar --assets Assets.zip
```

### Example 4: Hosted Mode Without External Auth

For testing or simple multi-instance setups:

```bash
export HSM_URL=http://localhost:8080

# Start HSM server (no auth)
hsm serve &

# Install and start Hytale server
./scripts/gsp-install-hytale-server.sh
./scripts/gsp-start-hytale-server.sh
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     HSM (Hytale Server Manager)              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐     ┌─────────────┐     ┌──────────────┐  │
│  │   CLI       │     │  HTTP API   │     │  Auth        │  │
│  │  Commands   │     │  Server     │     │  Middleware  │  │
│  │             │     │             │     │  (JWT/JWKS)  │  │
│  └──────┬──────┘     └──────┬──────┘     └──────┬───────┘  │
│         │                   │                    │          │
│         └───────────────────┼────────────────────┘          │
│                             │                               │
│                    ┌────────▼────────┐                      │
│                    │    Services     │                      │
│                    │  - Download     │                      │
│                    │  - Session      │                      │
│                    │  - Device Flow  │                      │
│                    └────────┬────────┘                      │
│                             │                               │
│                    ┌────────▼────────┐                      │
│                    │  Hytale Client  │                      │
│                    │  (API Client)   │                      │
│                    └─────────────────┘                      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## CI/CD

HSM includes GitHub Actions workflows for continuous integration and deployment.

### Workflows

| Workflow | Trigger | Description |
|----------|---------|-------------|
| **CI** | Push/PR to main | Runs tests, linting, and build verification |
| **Release** | Tag `v*` | Builds binaries for all platforms and creates GitHub release |
| **Docker** | Push to main, tags | Builds and pushes multi-arch Docker images to GHCR |

### Creating a Release

1. Tag the commit with a semantic version:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. The release workflow will automatically:
   - Build binaries for Linux (amd64, arm64), macOS (amd64, arm64), and Windows (amd64)
   - Generate SHA256 checksums
   - Create a GitHub release with all artifacts

### Docker Images

Docker images are automatically built and pushed to Docker Hub:

```bash
# Pull latest from main branch
docker pull yourusername/hsm:latest

# Pull specific version
docker pull yourusername/hsm:v1.0.0

# Pull by commit SHA
docker pull yourusername/hsm:abc1234
```

Multi-architecture images are built for `linux/amd64` and `linux/arm64`.

**Required Secrets:**
- `DOCKERHUB_USERNAME` - Your Docker Hub username
- `DOCKERHUB_TOKEN` - Docker Hub access token (create at https://hub.docker.com/settings/security)

## Security Considerations

- **Session Storage**: Session tokens are stored in `~/.config/hsm/session.json` by default
- **JWT Validation**: In hosted mode, all protected endpoints validate JWT tokens against the configured JWKS endpoint
- **Non-Root Container**: Docker container runs as non-root user (uid 1000)
- **Token Isolation**: In multi-user mode, each user's sessions are isolated by JWT subject claim

## License

[Add your license here]
