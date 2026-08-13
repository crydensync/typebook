package main

import (
	"net/http"

	"github.com/crydensync/cryden/v2"
)

type authHandlers struct {
	engine *cryden.Engine
}

// POST /api/signup
func (h *authHandlers) signUp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := cryden.SignUp(r.Context(), h.engine, req.Email, req.Password, callerIP(r))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"user_id": user.ID, "email": user.Email})
}

// POST /api/login
func (h *authHandlers) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokens, err := cryden.Login(r.Context(), h.engine, req.Email, req.Password, callerIP(r), userAgent(r))
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

// POST /api/refresh
func (h *authHandlers) refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokens, err := cryden.RefreshToken(r.Context(), h.engine, req.RefreshToken)
	if err != nil {
		// A reused/stolen token being presented here is exactly the
		// case CrydenSync's family-revocation exists for — surface it
		// as 401, the client must force a full re-login.
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

// POST /api/logout — requires auth (needs to know which user's
// session this is, and the session must belong to them)
func (h *authHandlers) logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := userIDFromContext(r)
	if err := cryden.Logout(r.Context(), h.engine, req.SessionID, userID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// POST /api/logout-all — requires auth
func (h *authHandlers) logoutAll(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	if err := cryden.LogoutAll(r.Context(), h.engine, userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out of all devices"})
}

// sessionDTO is what the client actually needs — deliberately excludes
// TokenHash and FamilyID. Even though TokenHash is a SHA-256 hash, not
// the raw refresh token, there's no reason to expose it over the API;
// unnecessary data exposure is worth avoiding on principle, not just
// when something is directly exploitable.
type sessionDTO struct {
	ID        string `json:"id"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	CreatedAt string `json:"created_at"`
}

// GET /api/sessions — requires auth
func (h *authHandlers) listSessions(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	sessions, err := cryden.ListSessions(r.Context(), h.engine, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]sessionDTO, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, sessionDTO{
			ID:        s.ID,
			IP:        s.IP,
			UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// DELETE /api/sessions/{id} — requires auth
func (h *authHandlers) revokeSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	userID := userIDFromContext(r)
	if err := cryden.RevokeSession(r.Context(), h.engine, sessionID, userID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "session revoked"})
}

// POST /api/change-password — requires auth
func (h *authHandlers) changePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := userIDFromContext(r)
	if err := cryden.ChangePassword(r.Context(), h.engine, userID, req.CurrentPassword, req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	// All sessions (including this one) were just revoked by
	// ChangePassword — the client must discard its tokens and
	// force a fresh login.
	writeJSON(w, http.StatusOK, map[string]string{"status": "password changed, please log in again"})
}

// POST /api/delete-account — requires auth
func (h *authHandlers) deleteAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CurrentPassword string `json:"current_password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := userIDFromContext(r)
	if err := cryden.DeleteAccount(r.Context(), h.engine, userID, req.CurrentPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "account deleted"})
}

// POST /api/email/request-change — requires auth
func (h *authHandlers) requestEmailChange(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NewEmail string `json:"new_email"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := userIDFromContext(r)
	if err := cryden.RequestEmailChange(r.Context(), h.engine, userID, req.NewEmail); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "verification email sent"})
}

// POST /api/email/confirm-change — public, the token itself is the
// proof of authorization, no Bearer auth needed (the user clicks a
// link from their email, they may not have an active session)
func (h *authHandlers) confirmEmailChange(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := cryden.ConfirmEmailChange(r.Context(), h.engine, req.Token); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "email changed"})
}