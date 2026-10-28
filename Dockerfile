# -------------------------
# 1) Builder stage
# -------------------------
FROM golang:1.24-alpine AS builder

ENV GOTOOLCHAIN=auto

WORKDIR /app
COPY . /app

RUN cd cmd/server && CGO_ENABLED=0 GOOS=linux go build -o ts6viewer .

# -------------------------
# 2) Runtime stage
# -------------------------
FROM alpine:latest

# Server-side timezone (neutral default: UTC) for the pre-formatted fallback
# strings and the log output. The frontend displays times in each viewer's
# own local timezone, so this only matters when JavaScript is disabled and
# for operator logs. Pick a zone per deployment with -e TZ=<zone>, e.g.
# -e TZ=Europe/Berlin or -e TZ=America/New_York.
ENV TZ=UTC

RUN apk add --no-cache ca-certificates gettext openssh-client tzdata

WORKDIR /app/cmd/server

COPY --from=builder /app /app

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
RUN chmod +x /app/cmd/server/ts6viewer

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
