# Typebook 📝

A minimalist note-taking app demonstrating **CrydenSync** authentication in action.

## ✨ Features

### Authentication (Powered by CrydenSync)
- ✅ User signup/login with JWT
- ✅ Rate limiting with headers
- ✅ SQLite persistence for users & sessions
- ✅ Session management (list active devices)
- ✅ Logout (single device) & Logout all devices
- ✅ Change password

### Notes
- ✅ Create notes (protected)
- ✅ List your notes
- ✅ Delete notes

## 🚀 Quick Start

```bash
# 1. Clone
git clone https://github.com/raymondproguy/typebook
cd typebook

# 2. Copy environment file
cp .env.example .env
# Edit .env with your JWT secret

# 3. Run
go mod tidy
go run main.go

# 4. Open http://localhost:3000
```

📡 API Endpoints

Method Endpoint Description Auth
POST /signup Create account ❌
POST /login Login + get tokens ❌
GET /health Health check ❌
POST /api/notes Create note ✅
GET /api/notes List notes ✅
DELETE /api/notes/:id Delete note ✅
POST /api/logout Logout device ✅
POST /api/logout-all Logout all devices ✅
POST /api/change-password Change password ✅
GET /api/sessions List sessions ✅

📝 Example Requests

Sign Up

```bash
curl -X POST http://localhost:3000/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"SecurePass123"}'
```

Login

```bash
curl -X POST http://localhost:3000/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"SecurePass123"}'
```

Create Note (with token)

```bash
curl -X POST http://localhost:3000/api/notes \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Meeting","content":"Discuss Typebook"}'
```

🏗️ Built With

· CrydenSync - Auth engine
· Fiber - Web framework
· SQLite - Database
· Go - Language

📊 Project Stats

 
⭐ Stars https://img.shields.io/github/stars/raymondproguy/typebook
📥 Downloads https://img.shields.io/github/downloads/raymondproguy/typebook/total
✅ Build https://github.com/raymondproguy/typebook/actions/workflows/test.yml/badge.svg

🎯 Purpose

This isn't just a demo — it's a real working app that shows how CrydenSync works in production. Perfect for:

· Learning JWT authentication flow
· Understanding session management
· Seeing rate limiting in action
· Building your own auth system

📚 Learn More

· CrydenSync Documentation
· Fiber Documentation
· Go SQLite

🤝 Contributing

PRs welcome! Feel free to add features or improvements.

📄 License

MIT © Raymond Nicholas

---

Built with ❤️ using CrydenSync


### **File 10: `docs/api.md`** (Detailed API docs)
```markdown
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
