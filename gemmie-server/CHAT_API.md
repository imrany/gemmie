# Chat API Documentation

This document describes the REST API endpoints for managing chats and messages in the Gemmie chatbot server.

## Authentication

All endpoints require the `X-User-ID` header containing the authenticated user's ID.

## Chat Endpoints

### Create Chat

Creates a new chat for the authenticated user.

- **URL**: `/api/chats`
- **Method**: `POST`
- **Headers**: 
  - `Content-Type: application/json`
  - `X-User-ID: {user_id}`

**Request Body:**
```json
{
  "title": "My New Chat"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Chat created successfully",
  "data": {
    "id": "chat_123",
    "user_id": "user_456",
    "title": "My New Chat",
    "created_at": "2023-12-01T10:00:00Z",
    "updated_at": "2023-12-01T10:00:00Z",
    "is_archived": false,
    "message_count": 0,
    "last_message_at": "2023-12-01T10:00:00Z"
  }
}
```

### Get All Chats

Retrieves all chats for the authenticated user.

- **URL**: `/api/chats`
- **Method**: `GET`
- **Headers**: 
  - `X-User-ID: {user_id}`

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Chats retrieved successfully",
  "data": [
    {
      "id": "chat_123",
      "user_id": "user_456",
      "title": "My Chat",
      "created_at": "2023-12-01T10:00:00Z",
      "updated_at": "2023-12-01T10:30:00Z",
      "is_archived": false,
      "message_count": 5,
      "last_message_at": "2023-12-01T10:30:00Z"
    }
  ]
}
```

### Get Single Chat

Retrieves a specific chat with its messages.

- **URL**: `/api/chats/{id}`
- **Method**: `GET`
- **Headers**: 
  - `X-User-ID: {user_id}`

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Chat retrieved successfully",
  "data": {
    "id": "chat_123",
    "user_id": "user_456",
    "title": "My Chat",
    "created_at": "2023-12-01T10:00:00Z",
    "updated_at": "2023-12-01T10:30:00Z",
    "is_archived": false,
    "message_count": 2,
    "last_message_at": "2023-12-01T10:30:00Z",
    "messages": [
      {
        "id": "msg_789",
        "chat_id": "chat_123",
        "role": "user",
        "content": "Hello!",
        "created_at": "2023-12-01T10:15:00Z",
        "model": ""
      },
      {
        "id": "msg_790",
        "chat_id": "chat_123",
        "role": "assistant",
        "content": "Hi! How can I help you today?",
        "created_at": "2023-12-01T10:30:00Z",
        "model": "gpt-3.5-turbo"
      }
    ]
  }
}
```

### Update Chat

Updates a chat's title or archive status.

- **URL**: `/api/chats/{id}`
- **Method**: `PUT`
- **Headers**: 
  - `Content-Type: application/json`
  - `X-User-ID: {user_id}`

**Request Body:**
```json
{
  "title": "Updated Chat Title",
  "is_archived": true
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Chat updated successfully",
  "data": {
    "id": "chat_123",
    "user_id": "user_456",
    "title": "Updated Chat Title",
    "created_at": "2023-12-01T10:00:00Z",
    "updated_at": "2023-12-01T11:00:00Z",
    "is_archived": true,
    "message_count": 2,
    "last_message_at": "2023-12-01T10:30:00Z"
  }
}
```

### Delete Chat

Deletes a chat and all its messages.

- **URL**: `/api/chats/{id}`
- **Method**: `DELETE`
- **Headers**: 
  - `X-User-ID: {user_id}`

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Chat deleted successfully"
}
```

## Message Endpoints

### Create Message

Adds a new message to a chat.

- **URL**: `/api/chats/{id}/messages`
- **Method**: `POST`
- **Headers**: 
  - `Content-Type: application/json`
  - `X-User-ID: {user_id}`

**Request Body:**
```json
{
  "role": "user",
  "content": "What's the weather like today?",
  "model": "gpt-3.5-turbo"
}
```

**Fields:**
- `role`: Either "user" or "assistant" (required)
- `content`: Message content (required)
- `model`: AI model used (optional, for assistant messages)

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Message created successfully",
  "data": {
    "id": "msg_791",
    "chat_id": "chat_123",
    "role": "user",
    "content": "What's the weather like today?",
    "created_at": "2023-12-01T11:00:00Z",
    "model": ""
  }
}
```

### Get Messages

Retrieves all messages for a specific chat.

- **URL**: `/api/chats/{id}/messages`
- **Method**: `GET`
- **Headers**: 
  - `X-User-ID: {user_id}`

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Messages retrieved successfully",
  "data": [
    {
      "id": "msg_789",
      "chat_id": "chat_123",
      "role": "user",
      "content": "Hello!",
      "created_at": "2023-12-01T10:15:00Z",
      "model": ""
    },
    {
      "id": "msg_790",
      "chat_id": "chat_123",
      "role": "assistant",
      "content": "Hi! How can I help you today?",
      "created_at": "2023-12-01T10:30:00Z",
      "model": "gpt-3.5-turbo"
    }
  ]
}
```

### Delete Message

Deletes a specific message.

- **URL**: `/api/messages/{id}`
- **Method**: `DELETE`
- **Headers**: 
  - `X-User-ID: {user_id}`

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Message deleted successfully"
}
```

## Error Responses

All endpoints may return the following error responses:

### 401 Unauthorized
```json
{
  "success": false,
  "message": "User ID header required"
}
```

### 403 Forbidden
```json
{
  "success": false,
  "message": "Access denied"
}
```

### 404 Not Found
```json
{
  "success": false,
  "message": "Chat not found"
}
```

### 400 Bad Request
```json
{
  "success": false,
  "message": "Invalid request body"
}
```

### 500 Internal Server Error
```json
{
  "success": false,
  "message": "Database error"
}
```

## Usage Examples

### Creating a complete conversation

1. **Create a new chat:**
```bash
curl -X POST http://localhost:8081/api/chats \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -d '{"title": "Weather Discussion"}'
```

2. **Add user message:**
```bash
curl -X POST http://localhost:8081/api/chats/chat_456/messages \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -d '{"role": "user", "content": "What'\''s the weather like?"}'
```

3. **Add assistant response:**
```bash
curl -X POST http://localhost:8081/api/chats/chat_456/messages \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -d '{"role": "assistant", "content": "I'\''d be happy to help with weather information! However, I don'\''t have access to real-time weather data.", "model": "gpt-3.5-turbo"}'
```

4. **Get the complete chat:**
```bash
curl -X GET http://localhost:8081/api/chats/chat_456 \
  -H "X-User-ID: user_123"
```

## Notes

- Messages are returned in chronological order (oldest first)
- Chat timestamps are automatically updated when messages are added
- Deleting a chat will also delete all associated messages
- The `message_count` field is automatically maintained when messages are added/removed
- Only the chat owner can view, modify, or delete chats and messages
- Chat titles default to "New Chat" if not provided
- The `last_message_at` timestamp is updated whenever a message is added to the chat