package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port		string
	DBPath		string
	JWTSecret	string
}

func InitEnv() AppConfig {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "flashcards.db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required.")
	}

	return AppConfig{
		Port:		port,
		DBPath: 	dbURL,
		JWTSecret: jwtSecret,
	}
}
