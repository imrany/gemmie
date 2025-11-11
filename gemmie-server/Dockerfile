# -------------------------
# Stage 1: Build the Go app
# -------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app/gemmie-server

# Install build deps
RUN apk add --no-cache git ca-certificates tzdata

# Copy go.mod and go.sum first (for dependency caching)
COPY gemmie-server/go.mod gemmie-server/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY gemmie-server/ .

# Build the binary
ARG VERSION=dev
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X github.com/imrany/gemmie/gemmie-server/internal/version.version=${VERSION}" \
    -o /gemmie-server ./main.go

# -------------------------
# Stage 2: Runtime image
# -------------------------
FROM alpine:3.19

# Install runtime deps
RUN set -eux && \
    apk add --no-cache ca-certificates tzdata && \
    addgroup -S gemmie && \
    adduser -S gemmie -G gemmie && \
    rm -rf /var/cache/apk/*

WORKDIR /opt/gemmie

# Copy binary from builder
COPY --from=builder /gemmie-server .

# Ensure executable
RUN chmod +x gemmie-server

# Run as non-root
USER gemmie

# Expose default port
EXPOSE 8081

# Default envs (override in docker-compose or runtime)
ENV WHATSAPP_DB_PATH=/var/opt/gemmie-server/whatsapp.db
ENV PORT=8081

LABEL org.opencontainers.image.title="gemmie-server" \
    org.opencontainers.image.description="gemmie's server side program." \
    org.opencontainers.image.source="https://github.com/imrany/gemmie/gemmie-server"

# Start the app
ENTRYPOINT ["./gemmie-server"]
