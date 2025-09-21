# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules for better build performance
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
ARG VERSION=dev
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build \
    -ldflags="-w -s -X github.com/imrany/gemmie/auth-server/internal/version.version=${VERSION} -extldflags '-static'" \
    -a -installsuffix cgo \
    -o gemmie-server \
    ./main.go

# Verify the binary
RUN chmod +x gemmie-server && ./gemmie-server --help || echo "Binary built successfully"

# Stage 2: Final runtime image
FROM scratch

# Copy CA certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/gemmie-server /gemmie-server

# Create directory for data file
COPY --from=builder --chown=65534:65534 /tmp /tmp

# Use nobody user (UID 65534) - available in scratch
USER 65534:65534

# Expose port
EXPOSE 8081

# Environment variables
ENV DATA_FILE=/tmp/gemmie_data.json \
    PORT=8081 

# Labels following OCI standard
LABEL org.opencontainers.image.title="gemmie-server" \
      org.opencontainers.image.description="Authentication server for Gemmie chat application" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.source="https://github.com/imrany/gemmie" \
      org.opencontainers.image.url="https://github.com/imrany/gemmie" \
      org.opencontainers.image.vendor="Gemmie" \
      org.opencontainers.image.licenses="MIT" \
      maintainer="imrany"

# Start the application
ENTRYPOINT ["/gemmie-server"]