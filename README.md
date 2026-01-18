# Hytale Server Manager (HSM)

A simple tool to download and manage your Hytale dedicated server.

## What Does It Do?

- Logs you into your Hytale account
- Downloads the Hytale server files for you
- Manages game sessions so your server can run

## Quick Start

### Option 1: Docker (Recommended)

Pull the pre-built image:

```bash
docker pull yourusername/hsm:latest
```

### Option 2: Download Binary

Download the latest release for your operating system from the [Releases Page](https://github.com/yourusername/hsm/releases).

Available for:
- Windows
- macOS (Intel & Apple Silicon)
- Linux (x64 & ARM)

---

## How to Use

### Step 1: Login to Hytale

**With Docker:**
```bash
docker run -it --rm -v $(pwd)/config:/home/hsm/.config yourusername/hsm login
```

**With Binary:**
```bash
hsm login
```

This will open your browser to log in with your Hytale account.

### Step 2: Download the Server

**With Docker:**
```bash
docker run -it --rm \
  -v $(pwd):/data \
  -v $(pwd)/config:/home/hsm/.config \
  yourusername/hsm download
```

**With Binary:**
```bash
hsm download
```

The server files will be downloaded and extracted to your current folder.

### Step 3: Get Game Session (for running the server)

**With Docker:**
```bash
docker run -it --rm \
  -v $(pwd)/config:/home/hsm/.config \
  -p 8080:8080 \
  yourusername/hsm serve
```

**With Binary:**
```bash
hsm serve
```

Then in another terminal, get your session tokens:

```bash
curl -X POST http://localhost:8080/game-session
```

This gives you the tokens needed to start your Hytale server.

---

## Common Options

| Option | Description |
|--------|-------------|
| `--output /path/to/folder` | Download server files to a specific folder |
| `--patchline prerelease` | Download prerelease version instead of stable |
| `--port 3000` | Run the HTTP server on a different port |

---

## Need Help?

- Check the [Releases Page](https://github.com/yourusername/hsm/releases) for the latest version
- Open an [Issue](https://github.com/yourusername/hsm/issues) if you run into problems

---

## For Advanced Users

<details>
<summary>Click to expand advanced documentation</summary>

### Running as a Service

You can run HSM as a background HTTP service using Docker Compose:

```yaml
services:
  hsm:
    image: yourusername/hsm:latest
    ports:
      - "8080:8080"
    volumes:
      - ./config:/home/hsm/.config
    command: serve
    restart: unless-stopped
```

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Check if HSM is running |
| `/download` | GET | Get server download URL |
| `/version` | GET | Get latest server version |
| `/game-session` | POST | Get game session tokens |

### For Hosting Providers (GSP Mode)

If you're a game server provider and need multi-user support with JWT authentication:

```bash
hsm serve --jwks-endpoint https://your-auth-server/.well-known/jwks.json
```

This enables JWT-based authentication where each user gets their own isolated session.

### Building from Source

```bash
git clone https://github.com/yourusername/hsm.git
cd hsm
go build -o hsm main.go
```

</details>

## License

[Add your license here]
