#!/bin/sh
set -eu

# Paths
EXAMPLE="/app/cmd/server/config.example.json"
TARGET_REL="config.json"                     # relative path in WORKDIR (/app/cmd/server)
TARGET_ABS="/app/cmd/server/config.json"     # absolute path for clarity
BINARY="/app/cmd/server/ts6viewer"

# Export variables used in the template (use environment variables only)
export API_KEY="${API_KEY:-}"
export SERVER_PORT="${SERVER_PORT:-8080}"
export BASE_URL="${BASE_URL:-http://127.0.0.1:10080}"
export SERVER_ID="${SERVER_ID:-1}"
export THEME="${THEME:-dark}"
export REFRESH_INTERVAL="${REFRESH_INTERVAL:-60}"

# Minimal, useful debug output
echo "=== STARTUP ==="
# Mask API key for logs: show first 4 chars if present
if [ -n "$API_KEY" ]; then
  echo "API_KEY=****${API_KEY#????}"
else
  echo "API_KEY=(empty)"
fi
echo "PORT=${SERVER_PORT}  BASE_URL=${BASE_URL}  THEME=${THEME}"
echo "Working dir: $(pwd)"
echo

# Quick binary check
if [ -x "$BINARY" ]; then
  echo "Binary: OK -> $BINARY"
else
  echo "Binary: MISSING or not executable -> $BINARY" >&2
fi

# Ensure example exists; if not, start the binary directly
if [ ! -f "$EXAMPLE" ]; then
  echo "config.example.json not found at $EXAMPLE, starting binary directly" >&2
  exec "$BINARY"
fi

# Generate config.json from config.example.json using envsubst
envsubst < "$EXAMPLE" > "$TARGET_ABS" || {
  echo "Failed to generate config.json" >&2
  exit 1
}

# Ensure relative copy exists for relative loads
cp -f "$TARGET_ABS" "$TARGET_REL" || true

# Show compact confirmation and a short preview of the generated config
size=$(wc -c < "$TARGET_ABS" 2>/dev/null || echo 0)
echo "Generated config.json (${size} bytes). Preview:"
echo "-----"
head -n 20 "$TARGET_ABS" || true
echo "-----"

echo "Starting server..."
exec "$BINARY"
