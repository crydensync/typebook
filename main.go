package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/raymondproguy/typebook/database"
	"github.com/raymondproguy/typebook/handlers"
	"github.com/raymondproguy/typebook/models"
)

func main() {
	// Initialize databases
	if err := database.Init(); err != nil {
		log.Fatal("❌ Failed to initialize databases:", err)
	}

	engine := database.AuthEngine
	noteStore := models.NewNoteStore(database.NotesDB)
	profileStore := models.NewProfileStore(database.NotesDB)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Typebook v1.0.0",
	})

	// CORS for production
	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("ALLOW_ORIGINS"),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	// Logger
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	// 🚀 Serve static files from public directory
	app.Use("/", filesystem.New(filesystem.Config{
		Root:   http.Dir("./public"),
		Browse: true,
		MaxAge: 3600,
	}))

	// API Routes
	api := app.Group("/api")

	// Public routes
	app.Post("/signup", handler.Signup(engine))
	app.Post("/login", handler.Login(engine))
	app.Get("/health", health)
	app.Get("/u/:username", handler.GetPublicProfile(profileStore))
	app.Get("/shared/:share_id", handler.GetSharedNote(noteStore))

	// Protected routes
	protected := api.Group("", handler.AuthMiddleware(engine))
	protected.Get("/profile", handler.GetProfile(profileStore))
	protected.Put("/profile", handler.UpdateProfile(profileStore))
	protected.Post("/notes", handler.CreateNote(noteStore))
	protected.Get("/notes", handler.ListNotes(noteStore))
	// protected.Get("/notes/:id", handler.GetNote(noteStore))
	protected.Put("/notes/:id", handler.UpdateNote(noteStore))
	protected.Post("/notes/:id/favorite", handler.ToggleFavorite(noteStore))
	protected.Post("/notes/:id/share", handler.ShareNote(noteStore))
	protected.Post("/notes/:id/unshare", handler.UnshareNote(noteStore))
	protected.Delete("/notes/:id", handler.DeleteNote(noteStore))
	protected.Get("/tags", handler.GetUserTags(noteStore))
	protected.Post("/logout", handler.Logout(engine))
	protected.Post("/logout-all", handler.LogoutAll(engine))
	protected.Post("/change-password", handler.ChangePassword(engine))
	protected.Get("/sessions", handler.ListSessions(engine))

	// Get host and port from env
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0" // Listen on all interfaces
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	addr := host + ":" + port

	// Graceful shutdown
	go func() {
		log.Printf("🚀 Typebook starting on http://%s", addr)
		log.Printf("📁 Frontend available at http://%s", addr)
		log.Printf("📚 API docs at http://%s/health", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down...")
	database.Close()
	app.Shutdown()
}

func health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":   "healthy",
		"auth":     "crydensync",
		"version":  "v1.0.0",
		"auth_db":  "sqlite",
		"notes_db": "sqlite",
		"frontend": "/",
	})
}
