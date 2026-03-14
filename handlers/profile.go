package handler

import (
    "time"
    
    "github.com/gofiber/fiber/v2"
    "github.com/raymondproguy/typebook/models"
)

type UpdateProfileRequest struct {
    DisplayName string `json:"display_name"`
    Username    string `json:"username"`
    Bio         string `json:"bio"`
    AvatarURL   string `json:"avatar_url"`
    Phone       string `json:"phone"`
    Location    string `json:"location"`
    Website     string `json:"website"`
}

// Get current user's profile
func GetProfile(store *models.ProfileStore) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(string)
        
        profile, err := store.GetByUserID(userID)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to get profile"})
        }
        
        // If no profile yet, return empty template
        if profile == nil {
            profile = &models.Profile{
                UserID:    userID,
                UpdatedAt: time.Now(),
            }
        }
        
        return c.JSON(profile)
    }
}

// Update profile
func UpdateProfile(store *models.ProfileStore) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(string)
        var req UpdateProfileRequest
        
        if err := c.BodyParser(&req); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
        }
        
        // Check if username is taken (if provided)
        if req.Username != "" {
            existing, err := store.GetByUsername(req.Username)
            if err != nil {
                return c.Status(500).JSON(fiber.Map{"error": "Failed to check username"})
            }
            if existing != nil && existing.UserID != userID {
                return c.Status(400).JSON(fiber.Map{"error": "Username already taken"})
            }
        }
        
        profile := &models.Profile{
            UserID:      userID,
            DisplayName: req.DisplayName,
            Username:    req.Username,
            Bio:         req.Bio,
            AvatarURL:   req.AvatarURL,
            Phone:       req.Phone,
            Location:    req.Location,
            Website:     req.Website,
            UpdatedAt:   time.Now(),
        }
        
        if err := store.Upsert(profile); err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to update profile"})
        }
        
        return c.JSON(profile)
    }
}

// Get profile by username (public)
func GetPublicProfile(store *models.ProfileStore) fiber.Handler {
    return func(c *fiber.Ctx) error {
        username := c.Params("username")
        
        profile, err := store.GetByUsername(username)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to get profile"})
        }
        if profile == nil {
            return c.Status(404).JSON(fiber.Map{"error": "User not found"})
        }
        
        // Return only public info
        return c.JSON(fiber.Map{
            "username":     profile.Username,
            "display_name": profile.DisplayName,
            "bio":         profile.Bio,
            "avatar_url":  profile.AvatarURL,
            "location":    profile.Location,
            "website":     profile.Website,
        })
    }
}
