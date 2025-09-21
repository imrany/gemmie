# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the binary
# Inject version from build arg (default: dev)
ARG VERSION=dev
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X github.com/imrany/gemmie/auth-server/internal/version.version=${VERSION}" \
    -o gemmie-server \
    ./main.go

# Stage 2: Final runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S gemmie-server && \
    adduser -S gemmie-server -G gemmie-server

WORKDIR /opt/gemmie-server

# Copy binary
COPY --from=builder /app/gemmie-server .

# Run as non-root
USER gemmie-server

# Expose default port
EXPOSE 8081

# Set defaults
ENV DATA_FILE=./gemmie_data.json \
    PORT=8081

LABEL org.opencontainers.image.title="gemmie-server" \
      org.opencontainers.image.description="gemmie's server side program." \
      org.opencontainers.image.source="https://github.com/imrany/gemmie/auth-server"

# Start the app
ENTRYPOINT ["./gemmie-server"]
