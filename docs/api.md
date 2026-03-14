# Typebook API Documentation

## Base URL
```

http://localhost:3000

```

## Authentication

All protected endpoints require a Bearer token in the Authorization header:
```

Authorization: Bearer <your_access_token>

```

## Endpoints

### Health Check
`GET /health`

Response:
```json
{
  "status": "healthy",
  "auth": "crydensync",
  "version": "v1.0.0",
  "database": "sqlite"
}
```

Sign Up

POST /signup

Request:

```json
{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

Response (201):

```json
{
  "message": "User created successfully",
  "user_id": "usr_123...",
  "email": "user@example.com"
}
```

Login

POST /login

Request:

```json
{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

Response (200):

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "usr_456...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

Headers:

```
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 4
X-RateLimit-Reset: 45
```

Create Note

POST /api/notes

Request:

```json
{
  "title": "Meeting Notes",
  "content": "Discuss Typebook features"
}
```

Response (201):

```json
{
  "id": "note_123...",
  "user_id": "usr_123...",
  "title": "Meeting Notes",
  "content": "Discuss Typebook features",
  "created_at": "2026-03-10T12:00:00Z"
}
```

List Notes

GET /api/notes

Response (200):

```json
[
  {
    "id": "note_123...",
    "user_id": "usr_123...",
    "title": "Meeting Notes",
    "content": "Discuss Typebook features",
    "created_at": "2026-03-10T12:00:00Z"
  }
]
```

Delete Note

DELETE /api/notes/:id

Response (200):

```json
{
  "message": "Note deleted"
}
```

Logout

POST /api/logout

Request:

```json
{
  "refresh_token": "usr_456..."
}
```

Response (200):

```json
{
  "message": "Logged out successfully"
}
```

Logout All Devices

POST /api/logout-all

Response (200):

```json
{
  "message": "Logged out from all devices"
}
```

Change Password

POST /api/change-password

Request:

```json
{
  "old_password": "SecurePass123",
  "new_password": "NewSecurePass456"
}
```

Response (200):

```json
{
  "message": "Password changed successfully"
}
```

List Active Sessions

GET /api/sessions

Response (200):

```json
{
  "sessions": [
    {
      "id": "sess_123...",
      "user_id": "usr_123...",
      "created_at": "2026-03-10T12:00:00Z",
      "expires_at": "2026-03-17T12:00:00Z"
    }
  ]
}
```

Error Responses

400 Bad Request

```json
{
  "error": "Invalid request body"
}
```

401 Unauthorized

```json
{
  "error": "Invalid or expired token"
}
```

404 Not Found

```json
{
  "error": "Note not found"
}
```

Rate Limiting

Login endpoint is rate limited to 5 attempts per minute per IP. Rate limit info is returned in headers:

· X-RateLimit-Limit: Maximum attempts
· X-RateLimit-Remaining: Attempts remaining
· X-RateLimit-Reset: Seconds until limit resets

## Profile Endpoints

### Get My Profile
`GET /api/profile`

Response:
```json
{
  "user_id": "usr_123...",
  "display_name": "Alice Wonder",
  "username": "alice",
  "bio": "Building cool stuff",
  "avatar_url": "https://example.com/avatar.jpg",
  "phone": "+2348012345678",
  "location": "Lagos, Nigeria",
  "website": "https://alice.dev",
  "updated_at": "2026-03-11T12:00:00Z"
}
```

Update Profile

PUT /api/profile

Request:

```json
{
  "display_name": "Alice Wonder",
  "username": "alice",
  "bio": "Building cool stuff with Go",
  "avatar_url": "https://example.com/avatar.jpg",
  "phone": "+2348012345678",
  "location": "Lagos, Nigeria",
  "website": "https://alice.dev"
}
```

View Public Profile

GET /u/:username

Response:

```json
{
  "username": "alice",
  "display_name": "Alice Wonder",
  "bio": "Building cool stuff with Go",
  "avatar_url": "https://example.com/avatar.jpg",
  "location": "Lagos, Nigeria",
  "website": "https://alice.dev"
}
```
