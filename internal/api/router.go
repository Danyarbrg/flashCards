package api

import (
	"net/http"

	"github.com/Danyarbrg/flashCards/internal/models"
	"github.com/gin-gonic/gin"
)

var flashcards []models.Flashcard
var idCounter int = 1

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Get all the cards.
	r.GET("/cards", func(c *gin.Context){
		c.JSON(http.StatusOK, flashcards)
	})

	// Input the new card.
	r.POST("/cards", func(c *gin.Context){
		var newCard models.Flashcard

		if err := c.ShouldBindJSON(&newCard); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		newCard.ID = idCounter
		idCounter++

		flashcards = append(flashcards, newCard)
		c.JSON(http.StatusCreated, newCard)
	})

	return r
}