package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/crydensync/cryden/v2"
	"github.com/crydensync/cryden/v2/store/postgres"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:5173" // Vite's default dev server port
	}
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open DB connection: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
	}

	// Real email delivery requires RESEND_API_KEY (and ideally EMAIL_FROM,
	// on a domain verified in your Resend account). Without it, fall back
	// to logging the link to the console — fine for local dev, but this
	// means confirmation links never actually reach real users, so this
	// must be set in any real deployment.
	var emailSender interface {
		SendVerification(ctx context.Context, to string, rawToken string) error
	}
	if resendKey := os.Getenv("RESEND_API_KEY"); resendKey != "" {
		fromAddress := os.Getenv("EMAIL_FROM")
		if fromAddress == "" {
			fromAddress = "onboarding@resend.dev" // Resend's test sender — only works until you verify your own domain
		}
		emailSender = &resendEmailSender{apiKey: resendKey, fromAddress: fromAddress, frontendURL: frontendURL}
		log.Printf("email delivery: Resend (from %s)", fromAddress)
	} else {
		emailSender = &consoleEmailSender{frontendURL: frontendURL}
		log.Printf("email delivery: console only (dev mode) — set RESEND_API_KEY for real delivery")
	}

	engine, err := cryden.New(cryden.Config{
		JWTSecret:      jwtSecret,
		Users:          postgres.NewUserStore(db),
		Sessions:       postgres.NewSessionStore(db),
		Audit:          postgres.NewAuditStore(db),
		Verifications:  postgres.NewVerificationStore(db),
		EmailSender:    emailSender,
		AccessTokenTTL: 15 * time.Minute,
	})
	if err != nil {
		log.Fatalf("failed to construct cryden engine: %v", err)
	}

	noteStore := NewNoteStore(db)

	auth := &authHandlers{engine: engine}
	notes := &noteHandlers{notes: noteStore}

	mux := http.NewServeMux()

	// Public auth endpoints
	mux.HandleFunc("POST /api/signup", auth.signUp)
	mux.HandleFunc("POST /api/login", auth.login)
	mux.HandleFunc("POST /api/refresh", auth.refresh)
	mux.HandleFunc("POST /api/email/confirm-change", auth.confirmEmailChange)

	// Authenticated account-management endpoints
	mux.HandleFunc("POST /api/logout", requireAuth(engine, auth.logout))
	mux.HandleFunc("POST /api/logout-all", requireAuth(engine, auth.logoutAll))
	mux.HandleFunc("GET /api/sessions", requireAuth(engine, auth.listSessions))
	mux.HandleFunc("DELETE /api/sessions/{id}", requireAuth(engine, func(w http.ResponseWriter, r *http.Request) {
		auth.revokeSession(w, r, r.PathValue("id"))
	}))
	mux.HandleFunc("POST /api/change-password", requireAuth(engine, auth.changePassword))
	mux.HandleFunc("POST /api/delete-account", requireAuth(engine, auth.deleteAccount))
	mux.HandleFunc("POST /api/email/request-change", requireAuth(engine, auth.requestEmailChange))

	// Notes — every route requires auth, this is the whole point
	mux.HandleFunc("GET /api/notes", requireAuth(engine, notes.list))
	mux.HandleFunc("POST /api/notes", requireAuth(engine, notes.create))
	mux.HandleFunc("PUT /api/notes/{id}", requireAuth(engine, func(w http.ResponseWriter, r *http.Request) {
		notes.update(w, r, r.PathValue("id"))
	}))
	mux.HandleFunc("DELETE /api/notes/{id}", requireAuth(engine, func(w http.ResponseWriter, r *http.Request) {
		notes.delete(w, r, r.PathValue("id"))
	}))

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	handler := withCORS(corsOrigin, mux)

	log.Printf("typebook backend listening on :%s (CORS origin: %s)", port, corsOrigin)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
