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
- **Statistics view** (toggle button in the header) with the **Online Clients graph** (TeamSpeak blue line + area, **24 h / 7 d / 14 d / 1 m** range buttons) and a **Recent Activity** log — the graph can instead be displayed on the main page below the server info (it is then never shown twice)
- **Badges** (ADMIN / VIP from server group names) and **country flags** on client rows
- Caching + rate‑limit protection  
- Pure **ServerQuery (SSH)** backend  
- Optional **voice status** (mute status, audio status, talking)  
- Statistics sections configurable (show/hide graph & activity list, entry limit, graph range + retention)  

---

## Online Clients Graph

The **Online Clients** graph is shown exactly once: by default it lives in
the **statistics view** (toggle button in the header). With
`show_on_main_page` set to `"true"` it is displayed on the main page below
the server info instead — the statistics view then only contains the
activity log, so the graph never appears twice.

- Line + area fill in TeamSpeak blue.
- The y‑axis always uses clean, round ticks (steps of 1/2/5 × 10ⁿ, integer ≥ 1).
- One sample per data poll (10 s dedupe); older samples are compacted into
  15-minute averages so a month of history stays small.
- **Range buttons** in the panel header: `24 h`, `7 d`, `14 d`, `1 m` — all
  available by default. The choice is remembered per browser (localStorage).
- The largest range is limited by the `graph_max_range` config value: smaller
  ranges always stay available as buttons, and the setting also controls how
  much history is stored (e.g. `7d` keeps only a week). Default: `1m`.

## Recent Activity

The statistics toggle in the header switches the channel tree to the
statistics view — the activity log, and the Online Clients graph as well when
it is not displayed on the main page (the toggle switches back). The log is
**one merged list**, not two groups:

- Online and recently offline clients **together**, sorted **newest first**
  (event time: connection start for online, last confirmed online for offline;
  nickname as tie-breaker).
- The **name and the timestamp are colored** — green = online, red =
  offline — so each entry reads like a colored log line.
- Every row shows an absolute date + time:
  `since DD.MM. HH:MM` / `last seen DD.MM. HH:MM`
  (same year: `DD.MM.YYYY HH:MM`). Never relative ("just now", "2 h ago").
- **Badges** (ADMIN / VIP, derived from server group names, admin wins) and
  **country flags** on every row.
- Limited to the last **10 entries by default** (`recent_activity_limit`,
  `0` = unlimited).

### Statistics Configuration

All statistics settings are optional — old `config.json` files keep working
unchanged. New keys live in the `statistics` block:

```json
"statistics": {
  "graph": "true",
  "graph_max_range": "1m",
  "show_on_main_page": "false",
  "recent_activity": "true",
  "recent_activity_limit": "10"
}
```

| Key | Meaning | Default |
| --- | --- | --- |
| `statistics.graph` | Enable the Online Clients graph in the statistics view (`true`/`false`) | `true` |
| `statistics.graph_max_range` | Largest graph range: `24h` \| `7d` \| `14d` \| `1m` (also sets history retention) | `1m` |
| `statistics.show_on_main_page` | Show the graph on the main page below the server info — it is then not shown again in the statistics view (`true`/`false`) | `false` |
| `statistics.recent_activity` | Show the Recent Activity log (`true`/`false`) | `true` |
| `statistics.recent_activity_limit` | Number of entries to show, `0` = unlimited | `10` |

Docker environment variables (same defaults): `STATS_GRAPH`,
`STATS_GRAPH_MAX_RANGE`, `STATS_SHOW_ON_MAIN_PAGE`, `STATS_RECENT_ACTIVITY`,
`STATS_RECENT_ACTIVITY_LIMIT`.

With **both** sections disabled the statistics toggle button disappears as
well — the viewer is then purely the channel tree.

### Persistence

The full statistics state (recent-activity log + graph history) is written to
`state.json` next to `config.json` after every data fetch (atomic write,
stale entries pruned) and reloaded at startup — so **a server restart does
not reset the log or the graph**; it continues from the last saved state.
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
- STATS_GRAPH (enable the Online Clients graph in the statistics view; default `true`)
- STATS_GRAPH_MAX_RANGE (`24h` | `7d` | `14d` | `1m`; default `1m`)
- STATS_SHOW_ON_MAIN_PAGE (also show the graph on the main page; default `false`)
- STATS_RECENT_ACTIVITY (show the Recent Activity log; default `true`)
- STATS_RECENT_ACTIVITY_LIMIT (entry count, `0` = unlimited; default `10`)

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
