package main

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type noteHandlers struct {
	notes *NoteStore
}

// GET /api/notes — requires auth
func (h *noteHandlers) list(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	notes, err := h.notes.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if notes == nil {
		notes = []Note{} // empty array, not null, in the JSON response
	}
	writeJSON(w, http.StatusOK, notes)
}

// POST /api/notes — requires auth
func (h *noteHandlers) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Color string `json:"color"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Color == "" {
		req.Color = "default"
	}

	id := uuid.NewString()
	userID := userIDFromContext(r)
	n := Note{ID: id, UserID: userID, Title: req.Title, Body: req.Body, Color: req.Color}

	if err := h.notes.Create(r.Context(), n); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	n.CreatedAt = time.Now()
	n.UpdatedAt = n.CreatedAt
	writeJSON(w, http.StatusCreated, n)
}

// PUT /api/notes/{id} — requires auth, ownership checked
func (h *noteHandlers) update(w http.ResponseWriter, r *http.Request, noteID string) {
	existing, err := h.notes.GetByID(r.Context(), noteID)
	if err != nil {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	// Ownership check — same principle as CrydenSync's session
	// ownership check: knowing an ID is not the same as owning it.
	if existing.UserID != userIDFromContext(r) {
		writeError(w, http.StatusForbidden, "not your note")
		return
	}

	var req struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Color string `json:"color"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Color == "" {
		req.Color = existing.Color
	}

	if err := h.notes.Update(r.Context(), noteID, req.Title, req.Body, req.Color); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DELETE /api/notes/{id} — requires auth, ownership checked
func (h *noteHandlers) delete(w http.ResponseWriter, r *http.Request, noteID string) {
	existing, err := h.notes.GetByID(r.Context(), noteID)
	if err != nil {
		writeError(w, http.StatusNotFound, "note not found")
		return
	}
	if existing.UserID != userIDFromContext(r) {
		writeError(w, http.StatusForbidden, "not your note")
		return
	}

	if err := h.notes.Delete(r.Context(), noteID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}