# -------------------------
# Stage 1: Build the Go app
# -------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app/gemmie-server

RUN apk add --no-cache git ca-certificates tzdata

COPY gemmie-server/go.mod gemmie-server/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY gemmie-server/ .

ARG VERSION=dev
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X github.com/imrany/gemmie/gemmie-server/internal/version.version=${VERSION}" \
    -o /gemmie-server ./main.go

# -------------------------
# Stage 2: Runtime image
# -------------------------
FROM alpine:3.19

RUN set -eux && \
    apk add --no-cache ca-certificates tzdata && \
    addgroup -S gemmie && \
    adduser -S gemmie -G gemmie && \
    mkdir -p /var/opt/gemmie-server && \
    chown -R gemmie:gemmie /var/opt/gemmie-server

WORKDIR /opt/gemmie

COPY --from=builder /gemmie-server .

RUN chmod +x gemmie-server

USER gemmie

EXPOSE 8081

ENV DATA_FILE=/var/opt/gemmie-server/gemmie_data.json \
    PORT=8081

ENTRYPOINT ["./gemmie-server"]
