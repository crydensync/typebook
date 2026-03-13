package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/raymondproguy/typebook/database"
    "github.com/raymondproguy/typebook/handlers"
)

func main() {
    // Initialize Cryden with SQLite
    engine, err := database.InitAuth()
    if err != nil {
        log.Fatal("❌ Failed to initialize auth:", err)
    }
    defer engine.Close()
    log.Println("✅ Auth engine initialized")

    // 2. Create Fiber app
    app := fiber.New(fiber.Config{
        AppName: "Typebook v1.0.0",
    })

    // Middleware
    app.Use(cors.New())
    app.Use(logger.New(logger.Config{
        Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
    }))

    // Health check
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status":    "healthy",
            "auth":      "crydensync",
            "version":   "v1.0.0",
            "database":  "sqlite",
        })
    })

    // Public auth routes
    app.Post("/signup", handlers.Signup(engine))
    app.Post("/login", handlers.Login(engine))

    // Protected routes
    api := app.Group("/api", handlers.AuthMiddleware(engine))
    //api.Post("/notes", handlers.CreateNote())
    //api.Get("/notes", handlers.ListNotes())
    //api.Delete("/notes/:id", handlers.DeleteNote())
    api.Post("/logout", handlers.Logout(engine))
    api.Post("/logout-all", handlers.LogoutAll(engine))
    api.Post("/change-password", handlers.ChangePassword(engine))
    api.Get("/sessions", handlers.ListSessions(engine))

    // Get port from env
    port := os.Getenv("PORT")
    if port == "" {
        port = "3000"
    }

    // Graceful shutdown
    go func() {
        log.Printf("🚀 Server starting on port %s", port)
        if err := app.Listen(":" + port); err != nil {
            log.Fatal(err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("🛑 Shutting down server...")
    if err := app.Shutdown(); err != nil {
        log.Fatal("Server shutdown failed:", err)
    }
}
