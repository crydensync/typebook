package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/crydensync/cryden"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var AuthEngine *cryden.Engine
var NotesDB *sql.DB

func Init() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

logFile, err := os.OpenFile("data/typebook.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    
    fileLogger := security.NewFileAuditLogger("data/typebook.log")
    AuthEngine.WithAuditLogger(fileLogger)
    
    log.Println("✅ File logger initialized at data/typebook.log")

	// Create data directory if it doesn't exist
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	// Initialize Auth Database
	authPath := os.Getenv("AUTH_DB_PATH")
	if authPath == "" {
		authPath = filepath.Join(dataDir, "auth.db")
	}

	// Ensure directory exists for custom path too
	if err := os.MkdirAll(filepath.Dir(authPath), 0755); err != nil {
		return err
	}

	engine, err := cryden.WithSQLite(authPath)
	if err != nil {
		return err
	}
	AuthEngine = engine

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret != "" {
		cryden.WithJWTSecret(engine, jwtSecret)
	}

	log.Println("✅ Auth database initialized at", authPath)

	// Initialize Notes Database
	notesPath := os.Getenv("NOTES_DB_PATH")
	if notesPath == "" {
		notesPath = filepath.Join(dataDir, "notes.db")
	}

	notesDB, err := sql.Open("sqlite3", notesPath)
	if err != nil {
		return err
	}

	// Test connection
	if err := notesDB.Ping(); err != nil {
		return err
	}

	// Create notes table
	notesTable := `
    CREATE TABLE IF NOT EXISTS notes (
        id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        tags TEXT DEFAULT '',
        favorite BOOLEAN DEFAULT 0,
        shared BOOLEAN DEFAULT 0,
        share_id TEXT UNIQUE,
        created_at DATETIME NOT NULL,
        updated_at DATETIME NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes(user_id);
    CREATE INDEX IF NOT EXISTS idx_notes_tags ON notes(tags);
    CREATE INDEX IF NOT EXISTS idx_notes_favorite ON notes(favorite);
    CREATE INDEX IF NOT EXISTS idx_notes_share_id ON notes(share_id);
    `

	if _, err := notesDB.Exec(notesTable); err != nil {
		return err
	}

	// Create profiles table
	profilesTable := `
    CREATE TABLE IF NOT EXISTS profiles (
        user_id TEXT PRIMARY KEY,
        display_name TEXT,
        username TEXT UNIQUE,
        bio TEXT,
        avatar_url TEXT,
        phone TEXT,
        location TEXT,
        website TEXT,
        updated_at DATETIME NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_profiles_username ON profiles(username);
    `

	if _, err := notesDB.Exec(profilesTable); err != nil {
		return err
	}

	NotesDB = notesDB
	log.Println("✅ Notes database initialized at", notesPath)

	return nil
}

func Close() {
	// if AuthEngine != nil {
	//      AuthEngine.Close()
	// }
	if NotesDB != nil {
		NotesDB.Close()
	}
}
