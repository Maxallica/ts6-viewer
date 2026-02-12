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

RUN apk add --no-cache ca-certificates gettext openssh-client

WORKDIR /app/cmd/server

COPY --from=builder /app /app

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
RUN chmod +x /app/cmd/server/ts6viewer

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
