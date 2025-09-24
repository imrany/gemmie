# Gemmie Authentication Server Setup Guide

## Overview

This implementation provides secure authentication and data synchronization across devices using:
- **Go backend**: Stores only hashed credentials and encrypted user data
- **JSON file storage**: Simple, portable data storage
- **Vue frontend**: Enhanced authentication with device sync

## Security Features

- ✅ **No plain text passwords**: Only SHA-256 hashes stored
- ✅ **Salted hashing**: Uses email as salt for password hashing
- ✅ **Device sync**: Automatic data synchronization across devices
- ✅ **Local storage**: Data cached locally for offline access
- ✅ **CORS enabled**: Secure cross-origin requests

## Backend Setup (Go Server)

### 1. Prerequisites
```bash
# Install Go (1.21 or later)
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
sudo mkdir -p ~/.gemmie-server
sudo touch ~/.gemmie-server/gemmie_data.json
sudo chmod 666 ~/.gemmie-server/gemmie_data.json
# you need to fix permissions so the container’s gemmie user can write
sudo chown -R 1000:1000 ~/.gemmie-server


docker run -d  --name gemmie-server  -p 8081:8081 -v ~/.gemmie-server:/var/opt/gemmie-server -e PORT=8081   -e DATA_FILE=/var/opt/gemmie-server/gemmie_data.json ghcr.io/imrany/gemmie-server:latest
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

```
auth-server/
├── CODEOWNERS
├── gemmie_data.json
├── go.mod
├── go.sum
├── internal
│   ├── encrypt
│   │   └── encrypt.go
│   └── handlers
│       └── handlers.go
├── LICENSE
├── main.go
├── Makefile
├── README.md
├── scripts
│   ├── docker-compose.yaml
│   └── Dockerfile
├── SECURITY.md
└── store
    └── store.go
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
3. **Rate limiting**: Add rate limiting for API endpoints
4. **File permissions**: Secure the JSON storage file
5. **Backup**: Regular backups of `gemmie_data.json`

### Environment Variables
- `PORT`: Server port (default: 8081)
- `DATA_FILE`: Server data file (default: `gemmie_data.json`)

## Troubleshooting

### Common Issues

1. **CORS errors**
   - Check API_BASE_URL in frontend
   - Verify server CORS configuration

2. **Sync not working**
   - Check network connection
   - Verify X-User-ID header is sent
   - Check browser console for errors

3. **Server won't start**
   - Verify Go installation: `go version`
   - Check port availability: `netstat -an | grep 8081`
   - Check file permissions for JSON storage

4. **Data not persisting**
   - Check write permissions in server directory
   - Verify `gemmie_data.json` is being created
   - Check server logs for errors

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

## Security Notes

- Passwords are hashed with SHA-256 + email salt
- No plain text passwords stored anywhere
- Session management through client-side tokens
- All user data encrypted in JSON storage
- Cross-device sync uses secure authentication
