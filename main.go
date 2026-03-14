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
	"github.com/raymondproguy/typebook/models"
)

func main() {
	// Initialize Cryden with SQLite
	// engine, err := database.Init()
	if err := database.Init(); err != nil {
		log.Fatal("❌ Failed to initialize auth:", err)
	}
	engine := database.AuthEngine
	//defer engine.Close()
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

	// Health ciheck
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":   "healthy",
			"auth":     "crydensync",
			"version":  "v1.0.0",
			"database": "sqlite",
		})
	})

	// Public auth routes
	app.Post("/signup", handler.Signup(engine))
	app.Post("/login", handler.Login(engine))

	// Protected routes
	api := app.Group("/api", handler.AuthMiddleware(engine))
	api.Post("/logout", handler.Logout(engine))
	api.Post("/logout-all", handler.LogoutAll(engine))
	api.Post("/change-password", handler.ChangePassword(engine))
	api.Get("/sessions", handler.ListSessions(engine))

	noteStore := models.NewNoteStore(database.NotesDB)
	api.Post("/notes", handler.CreateNote(noteStore))
	api.Get("/notes", handler.ListNotes(noteStore))
  api.Delete("/notes/:id", handler.DeleteNote(noteStore))


  profileStore := models.NewProfileStore(database.NotesDB)
  // Protected profiles routes:
  api.Get("/profile", handler.GetProfile(profileStore))
  api.Put("/profile", handler.UpdateProfile(profileStore))
   // Public profile
  app.Get("/u/:username", handler.GetPublicProfile(profileStore))

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
