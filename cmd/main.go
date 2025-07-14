package main

import (
	"log"

	"github.com/Danyarbrg/flashCards/internal/api"
	"github.com/Danyarbrg/flashCards/internal/config"
	"github.com/Danyarbrg/flashCards/internal/db"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.InitEnv()
	if err := db.InitDB(cfg.DBPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	router := api.SetupRouter()
	// Обслуживание статических файлов из папки public
	router.Static("/public", "./public")
	
	// Обслуживание HTML файлов
	router.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})
	
	router.GET("/cards.html", func(c *gin.Context) {
		c.File("./public/cards.html")
	})

	router.GET("/review", func(c *gin.Context) {
		c.File("./public/review.html")
	})

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}