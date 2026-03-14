package handler

import (
    "database/sql"
    
    "github.com/gofiber/fiber/v2"
    "github.com/raymondproguy/typebook/models"
)

type CreateNoteRequest struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}

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

        note, err := store.Create(userID, req.Title, req.Content)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to create note"})
        }

        return c.Status(201).JSON(note)
    }
}

func ListNotes(store *models.NoteStore) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(string)

        notes, err := store.ListByUser(userID)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to list notes"})
        }

        return c.JSON(notes)
    }
}

func DeleteNote(store *models.NoteStore) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(string)
        noteID := c.Params("id")

        err := store.Delete(userID, noteID)
        if err == sql.ErrNoRows {
            return c.Status(404).JSON(fiber.Map{"error": "Note not found"})
        }
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to delete note"})
        }

        return c.JSON(fiber.Map{"message": "Note deleted"})
    }
}
