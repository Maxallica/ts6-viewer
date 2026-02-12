#!/bin/sh
set -eu

EXAMPLE="/app/cmd/server/config.example.json"
TARGET_ABS="/app/cmd/server/config.json"
BINARY="/app/cmd/server/ts6viewer"

export SERVER_PORT="${SERVER_PORT:-8080}"
export THEME="${THEME:-dark}"
export REFRESH_INTERVAL="${REFRESH_INTERVAL:-60}"
export HOST_CONNECTION_LINK="${HOST_CONNECTION_LINK:-}"

export HOST="${HOST:-localhost}"
export PORT="${PORT:-10022}"
export USER="${USER:-serveradmin}"
export PASSWORD="${PASSWORD:-}"
export ENABLE_VOICE_STATUS="${ENABLE_VOICE_STATUS:-true}"
export SERVER_ID="${SERVER_ID:-1}"

echo "[entrypoint] starting TS6 Viewer"

if [ ! -x "$BINARY" ]; then
  echo "[entrypoint] ERROR: binary not found or not executable: $BINARY" >&2
  exit 1
fi

if [ ! -f "$EXAMPLE" ]; then
  echo "[entrypoint] WARNING: config.example.json not found, starting without generated config" >&2
  echo "[entrypoint] Starting server..."
  exec "$BINARY"
fi

if ! envsubst < "$EXAMPLE" > "$TARGET_ABS"; then
  echo "[entrypoint] ERROR: failed to generate config.json" >&2
  exit 1
fi

echo "[entrypoint] config.json generated"

echo "[entrypoint] Loaded configuration:"
echo "  SERVER_PORT=$SERVER_PORT"
echo "  THEME=$THEME"
echo "  REFRESH_INTERVAL=$REFRESH_INTERVAL"
echo "  HOST=$HOST"
echo "  PORT=$PORT"
echo "  USER=$USER"
echo "  PASSWORD=*********"
echo "  ENABLE_VOICE_STATUS=$ENABLE_VOICE_STATUS"
echo "  SERVER_ID=$SERVER_ID"

echo "[entrypoint] Starting server..."

exec "$BINARY"
