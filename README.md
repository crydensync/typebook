# typebook

A minimal Google Keep-style notes app — built to exercise every feature of [CrydenSync](https://github.com/crydensync/cryden) in a real, full-stack, production-shaped setting. This is the reference app proving the auth engine works end-to-end, not just in unit tests.

## What this actually exercises

Every CrydenSync feature has a real reason to run here, not a contrived one:

- SignUp / Login / VerifyToken — basic auth flow
- RefreshToken — the API client silently refreshes on a 401 and retries
- ListSessions / RevokeSession — Settings → Active sessions
- Logout / LogoutAll
- ChangePassword — confirm it actually revokes all sessions (you'll get logged out)
- RequestEmailChange / ConfirmEmailChange — full round trip via the `/confirm-email?token=...` route
- DeleteAccount — with cascade cleanup of sessions and notes

## Structure

```
typebook/
├── backend/     Go, net/http, wraps CrydenSync + a small notes domain
└── frontend/    React + Vite, light/dark theme, minimal Keep-style UI
```

## Running locally

**1. Database** — run CrydenSync's migration, then this repo's own notes migration, against your Postgres instance:
```bash
psql "$DATABASE_URL" -f path/to/cryden/store/postgres/migrations/0001_initial_schema.up.sql
psql "$DATABASE_URL" -f backend/migrations/0001_notes.up.sql
```

**2. Backend:**
```bash
cd backend
cp .env.example .env   # fill in DATABASE_URL and JWT_SECRET
go run .
```

**3. Frontend:**
```bash
cd frontend
cp .env.example .env   # defaults to http://localhost:8080, adjust if needed
npm install
npm run dev
```

Open the printed Vite URL (typically `http://localhost:5173`).

## Notes on this being a demo/reference app, not a template for your own production auth

- The `consoleEmailSender` in `backend/email.go` just logs the verification link — it's a dev stand-in for `notify.EmailSender`. Replace it with a real provider (SES, SendGrid, Postmark) before using this pattern in a real product.
- Tokens are stored in `localStorage` on the frontend for simplicity. A production app handling more sensitive data might prefer httpOnly cookies instead — this is a reasonable, common tradeoff for a notes app, not a universal recommendation.
- `Password validation` (`ValidateEmail`/`ValidatePassword`) is expected to be wired into the CrydenSync engine's `SignUp` call — confirm that's in place in your CrydenSync version before relying on this app's signup form to enforce a real password policy.

## License

MIT
