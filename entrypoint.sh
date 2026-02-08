#!/bin/sh
set -eu

EXAMPLE="/app/cmd/server/config.example.json"
TARGET_ABS="/app/cmd/server/config.json"
BINARY="/app/cmd/server/ts6viewer"

export SERVER_PORT="${SERVER_PORT:-8080}"
export THEME="${THEME:-dark}"
export REFRESH_INTERVAL="${REFRESH_INTERVAL:-60}"

export BASE_URL="${BASE_URL:-http://127.0.0.1:10080}"
export API_KEY="${API_KEY:-}"

export HOST="${HOST:-localhost}"
export PORT="${PORT:-10022}"
export USER="${USER:-serveradmin}"
export PASSWORD="${PASSWORD:-}"
export MODE="${MODE:-webquery}"
export ENABLE_DETAILED_CLIENT_INFO="${ENABLE_DETAILED_CLIENT_INFO:-true}"
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
echo "  BASE_URL=$BASE_URL"
echo "  API_KEY=$API_KEY"
echo "  HOST=$HOST"
echo "  PORT=$PORT"
echo "  USER=$USER"
echo "  PASSWORD=$PASSWORD"
echo "  MODE=$MODE"
echo "  ENABLE_DETAILED_CLIENT_INFO=$ENABLE_DETAILED_CLIENT_INFO"
echo "  SERVER_ID=$SERVER_ID"

echo "[entrypoint] Starting server..."

exec "$BINARY"
