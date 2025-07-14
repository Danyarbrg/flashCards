package main

import (
	"log"
	"fmt"

	"github.com/Danyarbrg/flashCards/internal/api"
	"github.com/Danyarbrg/flashCards/internal/config"
	"github.com/Danyarbrg/flashCards/internal/db"
)

func main() {
	cfg := config.InitEnv()
	fmt.Println(cfg.DBPath,"\n", cfg.Port)

	db.InitDB(cfg.DBPath)

	router := api.SetupRouter()

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
