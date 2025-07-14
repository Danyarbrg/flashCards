package api

import (
	"net/http"

	"github.com/Danyarbrg/flashCards/internal/models"
	"github.com/gin-gonic/gin"
)

// SetupRouter launches gin engine.
func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/cards", getFlashcards)
	r.POST("/cards", createFlashcard)

	return r
}

// getFlashcards outputs all cards from DB.
func getFlashcards(c *gin.Context) {
	cards, err := models.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB read error."})
		return
	}
	c.JSON(http.StatusOK, cards)
}

// createFlashcard processes POST, creating flashcard.
func createFlashcard (c *gin.Context) {
	var card models.Flashcard

	// Parsing JSON request into stuct.
	if err := c.ShouldBindJSON(&card); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect JSON."})
		return
	}

	// Saving into DB.
	if err := card.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Saving error."})
		return
	}

	c.JSON(http.StatusOK, card)
}