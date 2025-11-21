# Gemmie Server Setup Guide

## Overview

This implementation provides secure authentication and
data synchronization across devices.

## Backend Setup (Go Server)

### 1. Prerequisites

```bash
# Install Go (1.24.3 or later)
# https://golang.org/doc/install

# Verify installation
go version
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Run Server

```bash
# Development
go run main.go

# Or build and run
go build -o gemmie-server
./gemmie-server
```

```bash
docker run -d --name gemmie-server --env-file .env -p 8081:8081 ghcr.io/imrany/gemmie-server:latest
```

The server will start on `http://localhost:8081`

### 4. Environment Configuration

```bash
# Optional: Set custom port
export PORT=3001
go run main.go
```

## API Endpoints

### POST /api/register

Register a new user

```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword123"
}
```

### POST /api/login

Login existing user

```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword123"
}
```

### GET /api/sync

Get user data (requires X-User-ID header)

### POST /api/sync

Update user data (requires X-User-ID header)

```json
{
  "chats": "[{...}]",
  "link_previews": "{...}",
  "current_chat_id": "chat_123"
}
```

### GET /api/health

Health check endpoint

### POST /api/arcades

Create a new arcade (requires X-User-ID header)

```json
{
  "label": "My Arcade",
  "code": "console.log('Hello, world!');",
  "description": "A simple JavaScript arcade",
  "code_type": "javascript"
}
```

### GET /api/arcades/{id}

Get an arcade by ID (requires X-User-ID header)

### PUT /api/arcades/{id}

Update an arcade by ID (requires X-User-ID header)

```json
{
  "label": "Updated Arcade",
  "code": "console.log('Updated!');",
  "description": "An updated JavaScript arcade",
  "code_type": "javascript"
}
```

### DELETE /api/arcades/{id}

Delete an arcade by ID (requires X-User-ID header)

## Data Flow

### Registration/Login Flow

1. User enters credentials
2. Frontend sends to server
3. Server creates/validates hash
4. Server returns user data + synced content
5. Frontend stores locally and syncs

### Data Sync Flow

1. **Auto-sync**: Triggered every 5 minutes and on app focus
2. **Manual sync**: User clicks sync button
3. **Save sync**: Auto-triggered 2 seconds after data changes
4. **Logout sync**: Ensures data is synced before logout

### Cross-Device Experience

1. User logs in on Device A → Data synced from server
2. User makes changes → Auto-synced to server
3. User logs in on Device B → Gets latest data from server
4. Changes on Device B → Merged with existing data

## File Structure

```bash
gemmie-server/
├── CODEOWNERS
├── gemmie_data.json
├── go.mod
├── go.sum
├── internal
│   ├── encrypt
│   │   └── encrypt.go
├── internal
├── internal
│   ├── encrypt
│   │   └── encrypt.go
│   └── handlers
│       └── arcade.go
├── LICENSE
=======
│   ├── encrypt
│   │   └── encrypt.go
│   └── handlers
│       └── arcade.go
├── LICENSE
=======
│   └── handlers
│       └── arcade.go
=======
├── internal
│   ├── encrypt
│   │   └── encrypt.go
│   └── handlers
│       └── arcade.go
├── LICENSE
├── main.go
├── Makefile
├── README.md
├── SECURITY.md
├── scripts
│   ├── docker-compose.yaml
│   └── Dockerfile
└── store
    ├── arcade_ops.go
    ├── chat_ops.go
    ├── errors_ops.go
    ├── message_ops.go
    ├── migrations
    │   └── 000001_create_users_table.up.sql
    ├── migrations_ops.go
    ├── store.go
    ├── tranx_ops.go
    └── user_ops.go
```

## Production Deployment

### Backend Deployment

```bash
# Build for production
go build -o gemmie-server main.go

# Run with environment variables
export PORT=8081
./gemmie-server
```

### Security Considerations for Production

1. **HTTPS**: Use HTTPS in production
2. **CORS**: Restrict CORS origins to your frontend domains

### Environment Variables

- `PORT`: Server port (default: 8081)

## Troubleshooting

### Debug Mode

Add logging to your server:

```go
log.Printf("Request: %s %s", r.Method, r.URL.Path)
```

## Data Migration

If you have existing local data and want to migrate to the server:

1. **Export existing data** from localStorage
2. **Login/register** on the server
3. **Manual sync** will upload your local data
4. **Verify** data appears on other devices
