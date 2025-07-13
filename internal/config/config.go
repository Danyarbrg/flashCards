package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port   string
	DBPath string
}

func Init() AppConfig {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Cant find file .env.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")

	return AppConfig{
		Port:		port,
		DBPath: 	dbURL,
	}
}
