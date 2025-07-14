package main

import (
	"log"

	"github.com/Danyarbrg/flashCards/internal/api"
	"github.com/Danyarbrg/flashCards/internal/config"
	"github.com/Danyarbrg/flashCards/internal/db"
)

func main() {
	cfg := config.InitEnv()
	if err := db.InitDB(cfg.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	router := api.SetupRouter()
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}