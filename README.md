TS6 Viewer
<p align="center">
  <img src="dark.png" alt="Dark Theme" width="45%">
  <img src="light.png" alt="Light Theme" width="45%">
</p>

A lightweight, fast, and modern web viewer for **TeamSpeak 6** servers.  
TS6 Viewer connects to **ServerQuery (SSH, Port 10022)** and displays live server, channel, and client information.

The viewer is designed to be:

- Simple to deploy  
- Fast and lightweight  
- Fully client‑side auto‑refreshing  
- Compatible with any TS6 server  
- Customizable via a single `config.json` file  

---


## Features

- Live TeamSpeak 6 server viewer
- Auto‑refresh with configurable interval
- Dark and light themes (each web user selects their own via the theme button, with system default)
- Channel tree rendering with clients  
- Spacer and full‑width channel support  
- **Recent Activity** panel (toggle button in the header): online + recently offline clients, 24 h rolling
- **24 h online-count graph** in the Recent Activity panel (TeamSpeak blue line + area, hour axis)
- **Badges** (ADMIN / VIP from server group names) and **country flags** on client rows
- Caching + rate‑limit protection  
- Pure **ServerQuery (SSH)** backend  
- Optional **voice status** (mute status, audio status, talking)  

---

## Recent Activity

The **Recent Activity** button in the header switches the channel tree to the
statistics panel (the button toggles back). The panel shows:

- **24 h online-count graph** — line + area fill in TeamSpeak blue with an
  hour axis. One sample is taken per data poll (10 s dedupe), kept for 24 h.
- **Recently offline clients** — most recent first, red
  `last seen …` (the last time they were confirmed online).
- **Online clients** — most recently connected first, green
  `since HH:MM` (connection time).
- **Badges** (ADMIN / VIP, derived from server group names, admin wins) and
  **country flags** on every row.

All times are absolute (today: `HH:MM`, same year: `DD.MM. HH:MM`, earlier
year: `DD.MM.YYYY HH:MM`) — never relative ("just now", "2 h ago").

### Persistence

The full statistics state (recent-activity list + 24 h history) is written to
`state.json` next to `config.json` after every data fetch (atomic write,
stale entries pruned) and reloaded at startup — so **a server restart does
not reset the panel or the graph**; it continues from the last saved state.
The location can be overridden with the `STATE_FILE` environment variable.
In Docker, `compose.yml` already sets `STATE_FILE=/data/state.json` and
mounts a named volume at `/data`, so the state also survives container
recreation.

---

# ServerQuery (SSH)

TS6 Viewer communicates exclusively via ServerQuery over SSH.

Advantages:
- Persistent SSH connection
- Very fast, even with many clients
- clientlist -voice provides all audio/mute/talker info in a single call
- No REST overhead
- No per-client requests
- No rate limits

---

# Configuration Files in the Project

The repository includes two important configuration templates:

### `config.example.json`
This file is included in the project and serves as the template for your actual configuration.  
You can:

- rename it manually to `config.json`, or  
- let Docker generate `config.json` automatically using environment variables.

---

# Docker Support

This repository includes:
- Dockerfile.sh — multi‑stage build for Go + Alpine
- entrypoint.sh — generates config.json dynamically using environment variables
- compose.yml — ready to run the viewer with one command (incl. a named volume for the persisted statistics state)

This allows you to run TS6 Viewer fully containerized.

---

## Dockerfile.sh explained

The Dockerfile uses two stages:

### 1) Builder Stage
- Based on golang:1.20-alpine
- Copies the entire repository into /app
- Ensures config.example.json exists
- Normalizes go.mod to avoid Go version parsing issues
- Builds a static Linux binary: cmd/server/ts6viewer

### 2) Runtime Stage
- Based on alpine:latest
- Installs CA certificates + gettext (envsubst)
- Copies the built binary and assets from the builder
- Copies entrypoint.sh
- Exposes port 8080
- Starts the viewer via the entrypoint script

---

## entrypoint.sh explained

The entrypoint script:

1. Loads environment variables
2. Applies defaults if variables are missing
3. Uses envsubst to generate config.json from config.example.json
4. Starts the TS6 Viewer binary

Environment variables include:

- SERVER_PORT
- REFRESH_INTERVAL
- HOST
- PORT
- USER
- PASSWORD
- ENABLE_VOICE_STATUS
- SERVER_ID
- STATE_FILE (location of the persisted statistics state file; default `state.json` in the working directory)

This makes the Docker container fully configurable without editing files.

# compose.yml

The `compose.yml` file provides a straightforward example of how to configure and run the ts6viewer using Docker. It also shows how to switch between a locally built image and the official Docker Hub image:

```
# Use this if you built the image locally via:
# docker build -t ts6viewer:main .
# image: ts6viewer:main
# Use this if you want to pull the prebuilt image from Docker Hub
image: maxallica/ts6-viewer:latest
```

You can find the prebuilt Docker image here
[![Docker Pulls](https://img.shields.io/docker/pulls/maxallica/ts6-viewer.svg?logo=docker&label=pulls)](https://hub.docker.com/r/maxallica/ts6-viewer)

---


# Building the Program

The project is written in Go.

### Build on Linux

```sh
git clone https://github.com/Maxallica/ts6-viewer.git
cd ts6viewer
go build -o ts6viewer
```

### Build Windows binary on Linux
```sh
cd "*/ts6-viewer/cmd/server"
GOOS=windows GOARCH=amd64 go build -o ts6viewer.exe
```

### Running on Linux

```sh
./ts6viewer
```

### Build on Windows

```sh
git clone https://github.com/Maxallica/ts6-viewer.git
cd ts6viewer
go build -o ts6viewer.exe
```

### Build Linux binary on Windows

```sh
cd "*/ts6-viewer/cmd/server"
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o ts6viewer
```

### Running on Windows
```sh
.\ts6viewer.exe
```

---

## Navigate to the TS6 Viewer page

http(s)\/\/\<ip\>:\<port\>/ts6viewer

---

## ❤️ Made with love in Germany
