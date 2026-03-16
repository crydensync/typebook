package models

import (
	"database/sql"
	"time"
)

type Note struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      string    `json:"tags"`
	Favorite  bool      `json:"favorite"`
	Shared    bool      `json:"shared"`
	ShareID   string    `json:"share_id, omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NoteStore handles note database operations
type NoteStore struct {
	db *sql.DB
}

func NewNoteStore(db *sql.DB) *NoteStore {
	return &NoteStore{db: db}
}

func (s *NoteStore) GetDB() *sql.DB {
	return s.db
}

func (s *NoteStore) Create(userID, title, content string) (*Note, error) {
	note := &Note{
		ID:        generateID(),
		UserID:    userID,
		Title:     title,
		Content:   content,
		CreatedAt: time.Now(),
	}

	_, err := s.db.Exec(
		"INSERT INTO notes (id, user_id, title, content, created_at) VALUES (?, ?, ?, ?, ?)",
		note.ID, note.UserID, note.Title, note.Content, note.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return note, nil
}

func (s *NoteStore) ListByUser(userID string) ([]Note, error) {
	rows, err := s.db.Query(
		"SELECT id, user_id, title, content, created_at FROM notes WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (s *NoteStore) Delete(userID, noteID string) error {
	result, err := s.db.Exec("DELETE FROM notes WHERE id = ? AND user_id = ?", noteID, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func generateID() string {
	return time.Now().Format("20060102150405") + randomString(4)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(1)
	}
	return string(result)
}
