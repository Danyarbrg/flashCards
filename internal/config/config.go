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

var Cfg AppConfig

func Init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Cant find file .env.")
	}

	Cfg = AppConfig{
		Port:	os.Getenv("PORT"),
		DBPath:	os.Getenv("DATABASE_URL"),
	}
}
