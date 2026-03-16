package handler

import (
	"database/sql"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/raymondproguy/typebook/models"
)

type CreateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Tags    string `json:"tags"` // "work,idea,personal"
}

type UpdateNoteRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Tags     string `json:"tags"`
	Favorite *bool  `json:"favorite"` // Use pointer to detect if field was sent
}

// Create note
func CreateNote(store *models.NoteStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		var req CreateNoteRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		if req.Title == "" || req.Content == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Title and content are required",
			})
		}

		now := time.Now()
		note := &models.Note{
			ID:        generateID(),
			UserID:    userID,
			Title:     req.Title,
			Content:   req.Content,
			Tags:      req.Tags,
			Favorite:  false,
			Shared:    false,
			CreatedAt: now,
			UpdatedAt: now,
		}

		_, err := store.GetDB().Exec(
			`INSERT INTO notes (id, user_id, title, content, tags, favorite, shared, created_at, updated_at) 
             VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			note.ID, note.UserID, note.Title, note.Content, note.Tags,
			note.Favorite, note.Shared, note.CreatedAt, note.UpdatedAt,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create note"})
		}

		return c.Status(201).JSON(note)
	}
}

// List notes with filters
func ListNotes(store *models.NoteStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)

		// Get query params
		tag := c.Query("tag")
		favorite := c.Query("favorite") == "true"
		search := c.Query("q")

		query := `SELECT id, user_id, title, content, tags, favorite, shared, share_id, created_at, updated_at 
                  FROM notes WHERE user_id = ?`
		args := []interface{}{userID}

		if tag != "" {
			query += ` AND tags LIKE ?`
			args = append(args, "%"+tag+"%")
		}

		if favorite {
			query += ` AND favorite = 1`
		}

		if search != "" {
			query += ` AND (title LIKE ? OR content LIKE ?)`
			args = append(args, "%"+search+"%", "%"+search+"%")
		}

		query += ` ORDER BY favorite DESC, updated_at DESC`

		rows, err := store.GetDB().Query(query, args...)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to list notes"})
		}
		defer rows.Close()

		var notes []models.Note
		for rows.Next() {
			var n models.Note
			var shareID sql.NullString
			err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Content, &n.Tags,
				&n.Favorite, &n.Shared, &shareID, &n.CreatedAt, &n.UpdatedAt)
			if err != nil {
				continue
			}
			if shareID.Valid {
				n.ShareID = shareID.String
			}
			notes = append(notes, n)
		}

		return c.JSON(notes)
	}
}

// Update note
func UpdateNote(store *models.NoteStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		noteID := c.Params("id")

		var req UpdateNoteRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		// Build update query dynamically
		updates := []string{}
		args := []interface{}{}

		if req.Title != "" {
			updates = append(updates, "title = ?")
			args = append(args, req.Title)
		}

		if req.Content != "" {
			updates = append(updates, "content = ?")
			args = append(args, req.Content)
		}

		if req.Tags != "" {
			updates = append(updates, "tags = ?")
			args = append(args, req.Tags)
		}

		if req.Favorite != nil {
			updates = append(updates, "favorite = ?")
			args = append(args, *req.Favorite)
		}

		if len(updates) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "No fields to update"})
		}

		updates = append(updates, "updated_at = ?")
		args = append(args, time.Now())
		args = append(args, userID, noteID)

		query := `UPDATE notes SET ` + strings.Join(updates, ", ") +
			` WHERE user_id = ? AND id = ?`

		result, err := store.GetDB().Exec(query, args...)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update note"})
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Note not found"})
		}

		return c.JSON(fiber.Map{"message": "Note updated"})
	}
}

// Toggle favorite
func ToggleFavorite(store *models.NoteStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		noteID := c.Params("id")

		result, err := store.GetDB().Exec(
			`UPDATE notes SET favorite = NOT favorite, updated_at = ? 
             WHERE user_id = ? AND id = ?`,
			time.Now(), userID, noteID,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to toggle favorite"})
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Note not found"})
		}

		// Get updated note
		var favorite bool
		err = store.GetDB().QueryRow(
			`SELECT favorite FROM notes WHERE id = ?`, noteID,
		).Scan(&favorite)

		return c.JSON(fiber.Map{
			"message":  "Favorite toggled",
			"favorite": favorite,
		})
	}
}

// Share note (generate public link)
func ShareNote(store *models.NoteStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		noteID := c.Params("id")

		// Generate unique share ID
		shareID := generateID() + "_share"

		result, err := store.GetDB().Exec(
			`UPDATE notes SET shared = 1, share_id = ?, updated_at = ? 
             WHERE user_id = ? AND id = ?`,
			shareID, time.Now(), userID, noteID,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to share note"})
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Note not found"})
		}

		return c.JSON(fiber.Map{
			"message":   "Note shared",
			"share_url": "/shared/" + shareID,
		})
	}
}

// Unshare note
func UnshareNote(store *models.NoteStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		noteID := c.Params("id")

		result, err := store.GetDB().Exec(
			`UPDATE notes SET shared = 0, share_id = NULL, updated_at = ? 
             WHERE user_id = ? AND id = ?`,
			time.Now(), userID, noteID,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to unshare note"})
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Note not found"})
		}

		return c.JSON(fiber.Map{"message": "Note unshared"})
	}
}

// Get shared note (public - no auth)
func GetSharedNote(store *models.NoteStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		shareID := c.Params("share_id")

		var note models.Note
		var userID string
		err := store.GetDB().QueryRow(
			`SELECT id, user_id, title, content, tags, created_at 
             FROM notes WHERE share_id = ? AND shared = 1`,
			shareID,
		).Scan(&note.ID, &userID, &note.Title, &note.Content, &note.Tags, &note.CreatedAt)

		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "Shared note not found"})
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get note"})
		}

		// Don't expose user ID
		return c.JSON(fiber.Map{
			"title":      note.Title,
			"content":    note.Content,
			"tags":       note.Tags,
			"created_at": note.CreatedAt,
		})
	}
}

// Get all tags for user
func GetUserTags(store *models.NoteStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)

		rows, err := store.GetDB().Query(
			`SELECT tags FROM notes WHERE user_id = ? AND tags != ''`,
			userID,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get tags"})
		}
		defer rows.Close()

		tagSet := make(map[string]bool)
		for rows.Next() {
			var tags string
			rows.Scan(&tags)
			for _, tag := range strings.Split(tags, ",") {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tagSet[tag] = true
				}
			}
		}

		tags := make([]string, 0, len(tagSet))
		for tag := range tagSet {
			tags = append(tags, tag)
		}

		return c.JSON(fiber.Map{"tags": tags})
	}
}

// Delete note
func DeleteNote(store *models.NoteStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		noteID := c.Params("id")

		result, err := store.GetDB().Exec(
			"DELETE FROM notes WHERE id = ? AND user_id = ?",
			noteID, userID,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to delete note"})
		}

		rows, _ := result.RowsAffected()
		if rows == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "Note not found"})
		}

		return c.JSON(fiber.Map{"message": "Note deleted"})
	}
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
