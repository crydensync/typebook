package models

import (
	"database/sql"
	"time"
)

type Profile struct {
	UserID      string    `json:"user_id"`
	DisplayName string    `json:"display_name,omitempty"`
	Username    string    `json:"username,omitempty"`
	Bio         string    `json:"bio,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	Location    string    `json:"location,omitempty"`
	Website     string    `json:"website,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProfileStore struct {
	db *sql.DB
}

func NewProfileStore(db *sql.DB) *ProfileStore {
	return &ProfileStore{db: db}
}

// Create or update profile
func (s *ProfileStore) Upsert(profile *Profile) error {
	query := `
    INSERT INTO profiles (user_id, display_name, username, bio, avatar_url, phone, location, website, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(user_id) DO UPDATE SET
        display_name = excluded.display_name,
        username = excluded.username,
        bio = excluded.bio,
        avatar_url = excluded.avatar_url,
        phone = excluded.phone,
        location = excluded.location,
        website = excluded.website,
        updated_at = excluded.updated_at
    `

	_, err := s.db.Exec(query,
		profile.UserID, profile.DisplayName, profile.Username,
		profile.Bio, profile.AvatarURL, profile.Phone,
		profile.Location, profile.Website, time.Now(),
	)
	return err
}

// Get profile by user ID
func (s *ProfileStore) GetByUserID(userID string) (*Profile, error) {
	query := `SELECT user_id, display_name, username, bio, avatar_url, phone, location, website, updated_at 
              FROM profiles WHERE user_id = ?`

	var p Profile
	err := s.db.QueryRow(query, userID).Scan(
		&p.UserID, &p.DisplayName, &p.Username, &p.Bio,
		&p.AvatarURL, &p.Phone, &p.Location, &p.Website, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Profile not created yet
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Get profile by username
func (s *ProfileStore) GetByUsername(username string) (*Profile, error) {
	query := `SELECT user_id, display_name, username, bio, avatar_url, phone, location, website, updated_at 
              FROM profiles WHERE username = ?`

	var p Profile
	err := s.db.QueryRow(query, username).Scan(
		&p.UserID, &p.DisplayName, &p.Username, &p.Bio,
		&p.AvatarURL, &p.Phone, &p.Location, &p.Website, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
