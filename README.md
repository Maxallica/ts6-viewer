# TS6 Viewer
<p align="center">
  <img src="dark.png" alt="Dark Theme" width="45%">
  <img src="light.png" alt="Light Theme" width="45%">
</p>

A lightweight, fast, and modern web viewer for **TeamSpeak 6** servers.  
TS6 Viewer can connect to **WebQuery (REST API, Port 10011)** or **ServerQuery (SSH, Port 10022)** and display live server, channel, and client information.

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
- Dark and light themes  
- Channel tree rendering with clients  
- Spacer and full‑width channel support  
- Caching + rate‑limit protection  
- Supports **WebQuery (REST)** and **ServerQuery (SSH)**  
- Optional **detailed client info** (mute status, audio status, platform, idle time, etc.)  

---

# Configuration Files in the Project

The repository includes two important configuration templates:

### `config.example.json`
This file is included in the project and serves as the template for your actual configuration.  
You can:

- rename it manually to `config.json`, or  
- let Docker generate `config.json` automatically using environment variables.

### `docker-compose.yml`
A ready‑to‑use compose file is included in the project.  
You can start the viewer with:

```
docker compose up -d
```

Both files are designed so you can deploy the viewer quickly without editing the binary.

---

## Example `config.json`

```json
{
  "_comment": "TS6 Viewer Configuration File",
  "_comment2": "Rename this file to config.json and adjust the values to your setup.",

  "server_port": "${SERVER_PORT}",
  "_comment_server_port": "The port on which the TS6 Viewer web interface will be available. Usually '8080'",

  "theme": "${THEME}",
  "_comment_theme": "Choose between 'light' or 'dark' for the viewer theme. Usually 'dark'",

  "refresh_interval": "${REFRESH_INTERVAL}",
  "_comment_refresh_interval": "How often the viewer should auto-refresh (in seconds). Usually '60'",

  "teamspeak6": {
    "base_url": "${BASE_URL}",
    "_comment_base_url": "The WebQuery HTTP endpoint of your TeamSpeak 6 server. Usually http://<ip>:10011.",

    "api_key": "${API_KEY}",
    "_comment_api_key": "Your TeamSpeak 6 API key. It is shown ONCE in the server logs on first startup. If lost, create a new one.",

    "host": "${HOST}",
    "_comment_host": "The ServerQuery host. Usually the IP address of your TeamSpeak 6 server or 'localhost' if TS6 Viewer runs on the same machine.",

    "port": "${PORT}",
    "_comment_port": "The ServerQuery port. Usually 10022 for SSH.",

    "user": "${USER}",
    "_comment_user": "The ServerQuery user. Usually 'serveradmin' by default.",

    "password": "${PASSWORD}",
    "_comment_password": "The ServerQuery password. It is shown ONCE in the server logs on first startup.",

    "mode": "${MODE}",
    "_comment_mode": "Choose between 'webquery' [REST API, port 10011] or 'serverquery' [SSH, port 10022].",

    "enable_detailed_client_info": "${ENABLE_DETAILED_CLIENT_INFO}",
    "_comment_enable_detailed_client_info": "Whether to fetch detailed client info (mute status, audio status, etc.). Can increase load when using WebQuery.",

    "server_id": "${SERVER_ID}",
    "_comment_server_id": "The ID of the virtual server you want to display. Default is usually '1'."
  }
}
```

---

## WebQuery vs ServerQuery

### WebQuery (REST API, Port 10011)
- Simple HTTP requests  
- Limited client info unless `enable_detailed_client_info = true`  
- Requires one request per client for detailed info  
- Can cause load on large servers  
- Recommended for small servers or simple setups  

### ServerQuery (SSH, Port 10022)
- Persistent SSH connection  
- `clientlist -voice` provides all mute/audio/talker info in one call  
- Much faster for large servers  
- No REST rate limits  
- Recommended for medium/large servers  

---

## enable_detailed_client_info

### `"enable_detailed_client_info": true`
- WebQuery: performs `/clientinfo?clid=X` for every client  
- ServerQuery: uses `clientinfo` or extended `clientlist`  
- Shows:
  - Mic muted  
  - Output muted  
  - Talking  
  - Recording  
  - Platform  
  - Country  
  - Idle time  

### `"enable_detailed_client_info": false`
- Only basic client info is shown  
- No mute/audio icons  
- Much faster on WebQuery  

---

# Docker Support

The repository includes:

- `Dockerfile.sh` — multi‑stage build for Go + Alpine  
- `entrypoint.sh` — generates config.json dynamically using environment variables  
- `docker-compose.yml` — ready to run the viewer with one command  

This allows you to run TS6 Viewer fully containerized.

---

## Dockerfile.sh explained

The Dockerfile uses two stages:

### 1) Builder Stage
- Based on `golang:1.20-alpine`
- Copies the entire repository into `/app`
- Ensures `config.example.json` exists
- Normalizes `go.mod` to avoid Go version parsing issues
- Builds a static Linux binary: `cmd/server/ts6viewer`

### 2) Runtime Stage
- Based on `alpine:latest`
- Installs CA certificates + gettext (`envsubst`)
- Copies the built binary and assets from the builder
- Copies `entrypoint.sh`
- Exposes port 8080
- Starts the viewer via the entrypoint script

---

## entrypoint.sh explained

The entrypoint script:

1. Loads environment variables  
2. Applies defaults if variables are missing  
3. Uses `envsubst` to generate `config.json` from `config.example.json`  
4. Starts the TS6 Viewer binary  

Environment variables include:

- SERVER_PORT  
- THEME  
- REFRESH_INTERVAL  
- BASE_URL  
- API_KEY  
- HOST, PORT, USER, PASSWORD  
- MODE (`webquery` or `serverquery`)  
- ENABLE_DETAILED_CLIENT_INFO  
- SERVER_ID  

This makes the Docker container fully configurable without editing files.

---

# docker-compose.yml

A ready‑to‑use compose file is included in the project.  
You can start the viewer with:

```
docker compose up -d
```

Example:

```yaml
version: "3.9"

services:
  ts6viewer:
    image: ts6viewer:latest
    container_name: ts6viewer
    ports:
      - "9000:8080"

    environment:
      SERVER_PORT: "8080"
      THEME: "dark"
      REFRESH_INTERVAL: "60"

      BASE_URL: "http://192.168.178.195:10080"
      API_KEY: "secret_api_key"

      HOST: "192.168.178.195"
      PORT: "10022"
      USER: "serveradmin"
      PASSWORD: ""

      MODE: "webquery"

      ENABLE_DETAILED_CLIENT_INFO: "true"

      SERVER_ID: "1"

    restart: unless-stopped
```

---

# Building the Program

The project is written in Go.

### Build on Linux
```bash
git clone https://github.com/Maxallica/ts6-viewer.git
cd ts6viewer
go build -o ts6viewer
```

### Build Windows binary on Linux
```bash
cd "*/ts6-viewer/cmd/server"
GOOS=windows GOARCH=amd64 go build -o ts6viewer.exe
```

### Running on Linux
```bash
./ts6viewer
```

### Build on Windows
```bash
git clone https://github.com/Maxallica/ts6-viewer.git
cd ts6viewer
go build -o ts6viewer.exe
```

### Build Linux binary on Windows
```bash
cd "*/ts6-viewer/cmd/server"
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o ts6viewer
```

### Running on Windows
```bash
.\ts6viewer.exe
```

---

## Navigate to the TS6 Viewer page

```
http(s)://<ip>:<port>/ts6viewer
```

---

## ❤️ Made with love in Germany
