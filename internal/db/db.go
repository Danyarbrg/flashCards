package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3" // Регистрирует драйвер SQLite
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
		example TEXT
	);`

	if _, err = DB.Exec(createTable); err != nil {
		log.Fatalf("Creating table error: %v", err)
	}

	fmt.Println("DB connected and ready.")
} 