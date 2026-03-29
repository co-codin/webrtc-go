# webrtc-go

A real-time video chat application written in Go. Participants join a room for bi-directional video/audio, share a stream link so viewers can watch without a camera, and chat in real time — all via WebRTC and WebSocket.

## Features

- **Multi-peer rooms** — any number of participants send and receive video/audio through a server-side SFU
- **Stream viewer mode** — share a read-only link so viewers can watch without a camera or microphone
- **Real-time chat** — text chat for room participants and stream viewers (separate channels)
- **Live viewer count** — shows how many people are watching the stream
- **PWA-ready** — service worker caching, web app manifest, installable on mobile

## Tech stack

| Layer | Technology |
|-------|-----------|
| Server | Go 1.25, [Fiber v3](https://github.com/gofiber/fiber) |
| WebRTC | [pion/webrtc v3](https://github.com/pion/webrtc) — server-side SFU |
| WebSocket | [fasthttp/websocket](https://github.com/fasthttp/websocket) |
| Templates | Go `html/template` via gofiber/template |
| CSS | [Bulma](https://bulma.io) |
| TURN | [coturn](https://github.com/coturn/coturn) |
| Container | Docker, Docker Compose |

## Project layout

```
.
├── cmd/
│   └── main.go                 # entry point
├── internal/
│   ├── handlers/               # HTTP + WebSocket handlers
│   │   ├── room.go             # room page + WebRTC signaling
│   │   ├── stream.go           # stream page + viewer signaling
│   │   └── chat.go             # chat + viewer-count WebSocket handlers
│   └── server/
│       └── server.go           # Fiber app wiring, all routes
├── pkg/
│   ├── chat/
│   │   ├── hub.go              # broadcast hub
│   │   └── client.go           # per-connection client
│   └── webrtc/
│       ├── peers.go            # SFU: peer management, RTP forwarding, PLI
│       ├── room.go             # room registry + viewer count
│       └── stream.go           # stream room registry + viewer count
├── views/                      # HTML templates
│   ├── layouts/main.html
│   ├── partials/
│   ├── welcome.html
│   ├── peer.html               # room participant view
│   └── stream.html             # stream viewer view
├── assets/                     # static files (JS, CSS, fonts, SW)
├── containers/
│   ├── images/Dockerfile
│   └── composes/
│       ├── dc.dev.yml          # development compose
│       └── dc.prod.yml         # production compose
└── Makefile
```

## Quick start

### Local (no Docker)

```bash
# install dependencies
go mod download

# run the server
make run
# or: go run cmd/main.go --addr :8080
```

Open `http://localhost:8080`, click **Create Room**, and share the URL.

> **Note:** WebRTC peer connections require STUN/TURN. Locally, the built-in Google STUN server (`stun.l.google.com:19302`) is used. For connections across NAT you need the TURN server — see [Docker (dev)](#docker-dev) below.

### Docker (dev)

Builds the app image and starts a [coturn](https://github.com/coturn/coturn) TURN server alongside it.

```bash
make dev
```

The app is available at `http://localhost:8080`.

The TURN server listens on `localhost:3478`. To make the browser resolve the hostname used by the client JS (`turn.videochat`), add an entry to your `/etc/hosts`:

```
127.0.0.1  turn.videochat
```

Stop everything:

```bash
make dev-down
```

### Docker (prod)

```bash
make prod          # starts detached
make prod-logs     # follow logs
make prod-stop     # graceful stop
make prod-down     # remove containers
```

Override the default TURN credentials by creating a `.env` file next to the compose file, or exporting environment variables before running:

```bash
export TURN_USER=myuser
export TURN_PASS=s3cr3t
export TURN_REALM=example.com
make prod
```

## All make targets

```
make build        build binary to ./bin/webrtc-go
make run          run locally with go run
make test         run all tests with -race
make lint         go vet + staticcheck
make tidy         go mod tidy

make docker-build  build Docker image only
make dev           docker compose up --build (dev)
make dev-down      docker compose down (dev)
make dev-logs      tail dev logs

make prod          docker compose up -d --build (prod)
make prod-stop     graceful stop (prod)
make prod-down     docker compose down (prod)
make prod-logs     tail prod logs
```

## WebSocket API

All signaling uses JSON envelopes:

```json
{ "event": "<type>", "data": "<JSON string>" }
```

| Event | Direction | Payload |
|-------|-----------|---------|
| `offer` | server → client | `RTCSessionDescription` |
| `answer` | client → server | `RTCSessionDescription` |
| `candidate` | both | `RTCIceCandidateInit` |

### Endpoints

| Path | Purpose |
|------|---------|
| `GET /room/create` | Creates a room, redirects to `/room/:uuid` |
| `GET /room/:uuid` | Room participant page |
| `GET /room/:uuid/websocket` | WebRTC signaling (bidirectional) |
| `GET /room/:uuid/chat/websocket` | Chat messages |
| `GET /room/:uuid/viewer/websocket` | Live viewer count |
| `GET /stream/:suuid` | Stream viewer page |
| `GET /stream/:suuid/websocket` | WebRTC signaling (receive-only) |
| `GET /stream/:suuid/chat/websocket` | Viewer chat messages |
| `GET /stream/:suuid/viewer/websocket` | Live viewer count |

> The stream UUID equals the room UUID. The stream link for a room at `/room/abc` is `/stream/abc`.

## Running tests

```bash
make test
# or with verbose output:
go test -v -race ./...
```

Tests cover the chat hub, room/stream registry, viewer-count atomics, and HTTP handler responses.

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server listen port (overridden by `--addr` flag) |
| `TURN_USER` | `user` | TURN server username (prod compose only) |
| `TURN_PASS` | `password` | TURN server password (prod compose only) |
| `TURN_REALM` | `videochat` | TURN server realm (prod compose only) |

## TLS

Pass `--cert` and `--key` flags to enable HTTPS/WSS:

```bash
./bin/webrtc-go --addr :443 --cert server.crt --key server.key
```
