# -------------------------
# Stage 1: Build the Go app
# -------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build deps
RUN apk add --no-cache git ca-certificates tzdata

# Copy mod files first (better caching)
COPY go.mod go.sum ./

# Download deps with caching
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the binary with version injection
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

# Ensure it’s executable
RUN chmod +x gemmie-server

# Run as non-root
USER gemmie

# Expose default port
EXPOSE 8081

# Default envs (can be overridden at runtime)
ENV DATA_FILE=/opt/gemmie/gemmie_data.json \
    PORT=8081

LABEL org.opencontainers.image.title="gemmie-server" \
      org.opencontainers.image.description="gemmie's server side program." \
      org.opencontainers.image.source="https://github.com/imrany/gemmie/gemmie-server"

# Start the app
ENTRYPOINT ["./gemmie-server"]
