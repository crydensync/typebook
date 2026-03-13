package handlers

import (
    "strings"
    "github.com/gofiber/fiber/v2"
    "github.com/crydensync/cryden"
)

func AuthMiddleware(engine *cryden.Engine) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Get token from header
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(401).JSON(fiber.Map{
                "error": "No authorization token provided",
            })
        }

        // Extract Bearer token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return c.Status(401).JSON(fiber.Map{
                "error": "Invalid authorization format. Use: Bearer <token>",
            })
        }

        // Verify token with Cryden
        userID, err := cryden.VerifyToken(engine, parts[1])
        if err != nil {
            return c.Status(401).JSON(fiber.Map{
                "error": "Invalid or expired token",
            })
        }

        // Store user ID in context
        c.Locals("user_id", userID)
        return c.Next()
    }
}
