package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error

	// Connection to DB.
	if DB, err = sql.Open("sqlite3", dbPath); err != nil {
		log.Fatalf("DB connection error: %v", err)
		return err
	}

	// DB ping.
	if err = DB.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
		return err
	}

	// Creating users table.
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	);`
	if _, err = DB.Exec(createUsersTable); err != nil {
		log.Fatalf("Creating users table error: %v", err)
		return err
	}

	// Creating table flashcards with users_id
	createFlashcardsTable := `
	CREATE TABLE IF NOT EXISTS flashcards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		word TEXT NOT NULL,
		meaning TEXT NOT NULL,
		example TEXT,
		next_review DATETIME DEFAULT CURRENT_TIMESTAMP,
		interval INTEGER DEFAULT 1,
		repetitions INTEGER DEFAULT 0,
		ef REAL DEFAULT 2.5,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	CREATE INDEX IF NOT EXISTS idx_user_id ON flashcards(user_id);
	CREATE INDEX IF NOT EXISTS idx_next_review ON flashcards(next_review);
	`
	if _, err = DB.Exec(createFlashcardsTable); err != nil {
		log.Fatalf("Creating flashcards table error: %v", err)
		return err
	}

	log.Println("DB connected and ready.")
	return nil
}