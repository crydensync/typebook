package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNoteNotFound = errors.New("note not found")

type Note struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NoteStore struct {
	db *sql.DB
}

func NewNoteStore(db *sql.DB) *NoteStore {
	return &NoteStore{db: db}
}

func (s *NoteStore) Create(ctx context.Context, n Note) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO notes (id, user_id, title, body, color)
		VALUES ($1, $2, $3, $4, $5)
	`, n.ID, n.UserID, n.Title, n.Body, n.Color)
	return err
}

func (s *NoteStore) ListByUser(ctx context.Context, userID string) ([]Note, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, body, color, created_at, updated_at
		FROM notes WHERE user_id = $1
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.Color, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *NoteStore) GetByID(ctx context.Context, id string) (Note, error) {
	var n Note
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, title, body, color, created_at, updated_at
		FROM notes WHERE id = $1
	`, id).Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Color, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Note{}, ErrNoteNotFound
	}
	return n, err
}

func (s *NoteStore) Update(ctx context.Context, id, title, body, color string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE notes SET title = $1, body = $2, color = $3, updated_at = now()
		WHERE id = $4
	`, title, body, color, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoteNotFound
	}
	return nil
}

func (s *NoteStore) Delete(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoteNotFound
	}
	return nil
}