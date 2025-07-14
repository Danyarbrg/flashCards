package api

import (
	"net/http"
	"strconv"

	"github.com/Danyarbrg/flashCards/internal/models"
	"github.com/gin-gonic/gin"
)

// SetupRouter launches gin engine.
func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/cards", getFlashcards)
	r.POST("/cards", createFlashcard)
	r.DELETE("/cards/:id", deleteFlashcard)
	r.PUT("/cards/:id", updateFlashcard)
	r.GET("/cards/:id", getFlashcardByID)

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

// DeleteFlash delete card.
func deleteFlashcard(c *gin.Context) {
	idStr := c.Param("id")
	
	// Convert string into int.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect ID."})
		return
	}

	// Trying to delete card
	if err := models.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete error."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card has been deleted."})
}

// UpdateFlashcard updates flashcard.
func updateFlashcard(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect ID."})
		return
	}

	// Structure for reading request body.
	var input struct {
		Word    string `json:"word"`
		Meaning string `json:"meaning"`
		Example string `json:"example"`
	}


	// Reading request body JSON.
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect JSON format."})
		return
	}

	// Update data into DB.
	if err := models.Update(id, input.Word, input.Meaning, input.Example); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update card error."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card updated."})
}

func getFlashcardByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect ID"})
		return
	}

	card, err := models.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Card can't be find."})
		return
	}

	c.JSON(http.StatusOK, card)
}
