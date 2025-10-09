#!/bin/bash

set -e

echo "Deploying Gemmie server..."
docker stop gemmie-server 2>/dev/null || true
docker rm gemmie-server 2>/dev/null || true
docker rmi ghcr.io/imrany/gemmie-server 2>/dev/null || true
docker pull ghcr.io/imrany/gemmie-server:latest
docker run -d --name gemmie-server --env-file .env -p 8081:8081 -v ~/.gemmie-server:/var/opt/gemmie-server ghcr.io/imrany/gemmie-server:latest
echo "Deployment complete. Showing logs:"
docker logs gemmie-server