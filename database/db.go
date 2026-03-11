package database

import (
    "database/sql"
    "log"
    "os"
    
    "github.com/crydensync/cryden"
    "github.com/joho/godotenv"
    _ "github.com/mattn/go-sqlite3"
)

// AuthDB - managed by CrydenSync
var AuthEngine *cryden.Engine

// NotesDB - Typebook database
var NotesDB *sql.DB

func Init() error {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables")
    }

    // Initialize Auth Database (CrydenSync)
    authPath := os.Getenv("AUTH_DB_PATH")
    if authPath == "" {
        authPath = "data/auth.db"
    }
    
    engine, err := cryden.WithSQLite(authPath)
    if err != nil {
        return err
    }
    AuthEngine = engine

    // Set JWT secret from env
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret != "" {
        cryden.WithJWTSecret(engine, jwtSecret)
    }
    
    log.Println("✅ Auth database initialized at", authPath)

    // Initialize Notes Database (Typebook)
    notesPath := os.Getenv("NOTES_DB_PATH")
    if notesPath == "" {
        notesPath = "data/notes.db"
    }
    
    notesDB, err := sql.Open("sqlite3", notesPath)
    if err != nil {
        return err
    }
    
    // Create notes table if not exists
    createTableSQL := `
    CREATE TABLE IF NOT EXISTS notes (
        id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        created_at DATETIME NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes(user_id);
    `
    
    if _, err := notesDB.Exec(createTableSQL); err != nil {
        return err
    }
    
    NotesDB = notesDB
    log.Println("✅ Notes database initialized at", notesPath)
    
    return nil
}

func Close() {
    if AuthEngine != nil {
        AuthEngine.Close()
    }
    if NotesDB != nil {
        NotesDB.Close()
    }
}
