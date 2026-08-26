package main

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/crydensync/cryden/v2"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// callerIP extracts the real client IP. Checks X-Forwarded-For first
// (set by reverse proxies/load balancers in real deployments), falls
// back to the raw connection address. This is exactly the kind of
// context-extraction the engine deliberately never does itself —
// it's the HTTP layer's job, per CrydenSync's framework-agnostic design.
func callerIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func userAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// requireAuth is middleware that verifies the Bearer access token and
// injects the authenticated user ID into the request context. Every
// notes endpoint and every account-management endpoint (change
// password, delete account, etc.) is wrapped by this.
func requireAuth(engine *cryden.Engine, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing or malformed Authorization header")
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := cryden.VerifyToken(engine, token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next(w, r.WithContext(ctx))
	}
}

func userIDFromContext(r *http.Request) string {
	id, _ := r.Context().Value(userIDContextKey).(string)
	return id
}

// withCORS allows the local Vite dev server (and, in prod, your
// deployed frontend origin) to call this API. Adjust allowedOrigin
// for your actual deployment.
func withCORS(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// Required for the OAuth link-init cookie handoff — a
		// credentialed cross-origin fetch() needs this to set/send
		// cookies. Safe alongside a specific allowedOrigin (never
		// combine with a wildcard "*" origin).
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
