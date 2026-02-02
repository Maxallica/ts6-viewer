# -------------------------
# 1) Builder stage
# -------------------------
FROM golang:1.20-alpine AS builder

# Install CA certificates required for building
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy local repository into the build image (portable across OSes)
COPY . /app

# Ensure config.example exists; keep a fallback copy
RUN cp cmd/server/config.example.json cmd/server/config.json || true

# Normalize go.mod 'go' directive to a valid major.minor form to avoid build errors
RUN if [ -f go.mod ]; then \
      sed -E -i 's/^go[[:space:]]+[0-9]+(\.[0-9]+){1,2}$/go 1.20/' go.mod || true; \
    fi

# Build a static Linux binary for the server
RUN cd cmd/server && CGO_ENABLED=0 GOOS=linux go build -o ts6viewer .

# -------------------------
# 2) Runtime stage
# -------------------------
FROM alpine:latest

# Install CA certificates and gettext for envsubst
RUN apk add --no-cache ca-certificates gettext

# Set working directory to where the server expects config.json when using relative path
WORKDIR /app/cmd/server

# Copy the built app and assets from builder
COPY --from=builder /app /app

# Copy entrypoint script and ensure it's executable
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh || true
RUN chmod +x /app/cmd/server/ts6viewer || true

# Expose the internal port the app listens on
EXPOSE 8080

# Use the entrypoint to generate config and start the app
ENTRYPOINT ["/app/entrypoint.sh"]
