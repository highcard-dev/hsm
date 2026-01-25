# Hytale Server Manager (HSM)

A simple tool to download and manage your Hytale dedicated server.

## What Does It Do?

- Logs you into your Hytale account
- Downloads the Hytale server files for you
- Manages game sessions so your server can run

## Quick Start

Download the latest release for your operating system from the [Releases Page](https://github.com/highcard-dev/hsm/releases).

Available for:

- Windows
- macOS (Intel & Apple Silicon)
- Linux (x64 & ARM)

Alternativly, you can also run HSM as a docker container

```bash
docker run highcard/hsm:latest -v $PWD:/data download
```

---

## Usage - Single User Mode

### Login to Hytale

```bash
hsm login
```

This will open your browser to log in with your Hytale account.

### Download the Server

**With Binary:**

```bash
hsm download
```

### Start the Server

**With Binary:**

```bash
hsm start
```

**With Docker:**

```bash
docker run -it --rm \
  -v $(pwd):/data \
  -v $(pwd)/config:/home/hsm/.config \
  highcard/hsm download
```

**Docker runs will ask for login credentials automatically, you don't need to call login manually.**

The server files will be downloaded and extracted to your current folder.

## Usage - Game Hosting providers

### Authentication

For Game Hosting Providers it is highly recommended to deploy HSM as a service to your infrastructure.
To manage the retrieval of game sessions, download URLs and anything else, a JWKS/JWT authentication flow can be used.
HSM Service will automatically secure every endpoint when you run it with the `--jwks-endpoint` flag (e.g., `hsm serve --jwks-endpoint https://your-auth-server/.well-known/jwks.json`).

For an example, take a look at the [hosted-auth example](examples/hosted-auth).

**This works very well with Kubernetes Service accounts too and is the way how it is used at druid.gg**

### No Authentication

If you disable authentication, make sure the service is not reachable from the outside world or by any entity (including your customers).
Otherwise someone can generate unlimited game sessions through your account.
Depending on your setup, authentication can be omitted if the customer does not have enough permission to abuse the session generation.
This highly depends on your exact setup!

Checkout the no-auth [hosted-no-auth example](examples/hosted-no-auth).

### Run HSM as Service

#### Helm Chart

Add this repository to Helm:

```
helm repo add hsm https://highcard-dev.github.io/hsm/
helm repo update
```

Install the chart:

```
helm install hsm hsm/hsm
```

#### Docker Container

```bash
docker run -it --rm \
  -v $(pwd)/config:/home/hsm/.config \
  -p 8080:8080 \
  highcard/hsm serve
```

#### Binary

```bash
hsm serve
```

When no session.json is found, use the link in the console to authenticate yourself.

### Retreive download url for latest game version

```bash
curl -X POST http://localhost:8080/download
```

Returns a presigned URL for the serverfile archive.

### Get Game Session

```bash
curl -X POST http://localhost:8080/game-session
```

This gives you the tokens needed to start your Hytale server.

### REST API

TODO: Readme
