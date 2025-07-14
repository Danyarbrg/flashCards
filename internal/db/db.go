package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) {
	var err error

	// Connection to DB.
	if DB, err = sql.Open("sqlite3", dbPath); err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	// DB ping.
	if err = DB.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}

	// Creating table
	createTable := `
	CREATE TABLE IF NOT EXISTS flashcards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word TEXT NOT NULL,
		meaning TEXT NOT NULL,
		example TEXT,
		next_review DATETIME DEFAULT CURRENT_TIMESTAMP,
		interval INTEGER DEFAULT 1,
		repetitions INTEGER DEFAULT 0,
		ef REAL DEFAULT 2.5
	);`

	if _, err = DB.Exec(createTable); err != nil {
		log.Fatalf("Creating table error: %v", err)
	}

	log.Println("DB connected and ready.")
} 