package handler

import (
	"context"
	"os"
	"strconv"

	"github.com/crydensync/cryden"
	"github.com/gofiber/fiber/v2"
)

var ctx = context.Background()

type SignupRequst struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SinginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"` 
}

type ChangePassword struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// Signup handler
func Signup(engine *cryden.Engine) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req SignupRequest
        if err := c.BodyParser(&req); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
        }

        user, err := cryden.SignUp(ctx, engine, req.Email, req.Password)
        if err != nil {
            return c.Status(400).JSON(fiber.Map{"error": err.Error()})
        }

        return c.Status(201).JSON(fiber.Map{
            "message": "User created successfully",
            "user_id": user.ID,
            "email":   user.Email,
        })
    }
}

// Login handler
func Login(engine *cryden.Engine) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req LoginRequest
        if err := c.BodyParser(&req); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
        }

        tokens, rate, err := cryden.Login(ctx, engine, req.Email, req.Password)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{"error": err.Error()})
        }

        // Set rate limit headers
        c.Set("X-RateLimit-Limit", strconv.Itoa(rate.Limit))
        c.Set("X-RateLimit-Remaining", strconv.Itoa(rate.Remaining))
        c.Set("X-RateLimit-Reset", strconv.FormatInt(int64(rate.Reset.Seconds()), 10))

        return c.JSON(fiber.Map{
            "access_token":  tokens.AccessToken,
            "refresh_token": tokens.RefreshToken,
            "token_type":    "Bearer",
            "expires_in":    tokens.ExpiresIn,
        })
    }
}

// Logout handler
func Logout(engine *cryden.Engine) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var req LogoutRequest
        if err := c.BodyParser(&req); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
        }

        if err := cryden.Logout(ctx, engine, req.RefreshToken); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": err.Error()})
        }

        return c.JSON(fiber.Map{"message": "Logged out successfully"})
    }
}

// Logout all devices
func LogoutAll(engine *cryden.Engine) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(string)
        
        if err := cryden.LogoutAll(ctx, engine, userID); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": err.Error()})
        }

        return c.JSON(fiber.Map{"message": "Logged out from all devices"})
    }
}

// Change password
func ChangePassword(engine *cryden.Engine) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(string)
        var req ChangePasswordRequest

        if err := c.BodyParser(&req); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
        }

        if err := cryden.ChangePassword(ctx, engine, userID, req.OldPassword, req.NewPassword); err != nil {
            return c.Status(400).JSON(fiber.Map{"error": err.Error()})
        }

        return c.JSON(fiber.Map{"message": "Password changed successfully"})
    }
}

// List active sessions
func ListSessions(engine *cryden.Engine) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(string)
        
        sessions, err := cryden.ListSessions(ctx, engine, userID)
        if err != nil {
            return c.Status(400).JSON(fiber.Map{"error": err.Error()})
        }

        return c.JSON(fiber.Map{"sessions": sessions})
    }
}
