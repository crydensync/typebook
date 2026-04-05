# Typebook 📝

The Officail CrydenSync Demo App.

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
git clone https://github.com/crydensync/typebook
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

## 👤 User Profiles

Typebook now supports user profiles! Each user can have:

- Display name
- Unique username
- Bio
- Avatar URL
- Phone number
- Location
- Website

### Profile Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/profile` | Get your profile | ✅ |
| PUT | `/api/profile` | Update profile | ✅ |
| GET | `/u/:username` | View public profile | ❌ |

### Example: Update Profile

```bash
curl -X PUT http://localhost:3000/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Alice Wonder",
    "username": "alice",
    "bio": "Building cool stuff with Go",
    "location": "Lagos, Nigeria",
    "website": "https://alice.dev"
  }'
```

Example: View Public Profile

```bash
curl http://localhost:3000/u/alice
```

### 🔍 Search Notes
```bash
# Search in titles and content
GET /api/notes/search?q=golang
GET /api/notes/search?tag=work
GET /api/notes/search?favorite=true
```

⭐ Favorites

```bash
# Mark important notes
POST /api/notes/:id/favorite  # Toggle favorite status
GET /api/notes?favorite=true   # Show only favorites
```

🔗 Note Sharing

```bash
# Share notes publicly
POST /api/notes/:id/share     # Generate share link
POST /api/notes/:id/unshare   # Remove sharing
GET /shared/:share_id          # Public view (no login)
```

🏷️ Tags & Categories

```bash
# Organize notes
POST /api/notes -d '{"tags": "work,idea,personal"}'
GET /api/notes?tag=work        # Filter by tag
GET /api/tags                  # List all your tags
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

MIT © CrydenSync

---

Built with ❤️ using CrydenSync
