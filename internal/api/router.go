package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Danyarbrg/flashCards/internal/config"
	"github.com/Danyarbrg/flashCards/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/register", register)
	r.POST("/login", login)

	protected := r.Group("/cards")
	protected.Use(AuthMiddleware())
	{
		protected.GET("", getFlashcards)
		protected.POST("", createFlashcard)
		protected.DELETE("/:id", deleteFlashcard)
		protected.PUT("/:id", updateFlashcard)
		protected.GET("/:id", getFlashcardByID)
		protected.GET("/due", getDueFlashcards)
		protected.POST("/review/:id", reviewFlashcard)
	}

	return r
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Удаляем префикс "Bearer "
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		cfg := config.InitEnv()
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id in token"})
			c.Abort()
			return
		}

		c.Set("user_id", int(userID))
		c.Next()
	}
}

func register(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	user, err := models.RegisterUser(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered",
		"user":    user,
	})
}

func login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	user, err := models.AuthenticateUser(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cfg := config.InitEnv()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
	})
}

func getFlashcards(c *gin.Context) {
	userID, _ := c.Get("user_id")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	sortBy := c.DefaultQuery("sort", "created")
	order := c.DefaultQuery("order", "asc")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}

	asc := order != "desc"
	offset := (page - 1) * limit

	cards, err := models.GetSortedPaginated(userID.(int), limit, offset, sortBy, asc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to read flashcards: %v", err)})
		return
	}
	c.JSON(http.StatusOK, cards)
}

func createFlashcard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var card models.Flashcard
	if err := c.ShouldBindJSON(&card); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if card.Word == "" || card.Meaning == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Word and meaning are required"})
		return
	}

	card.UserID = userID.(int)
	exists, err := models.ExistsByWord(userID.(int), card.Word)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to check word existence: %v", err)})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Word already exists for this user"})
		return
	}

	if err := card.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save flashcard: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Flashcard created",
		"card":    card,
	})
}

func deleteFlashcard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := models.Delete(id, userID.(int)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete flashcard: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Flashcard deleted"})
}

func updateFlashcard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input struct {
		Word    string `json:"word"`
		Meaning string `json:"meaning"`
		Example string `json:"example"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if input.Word == "" || input.Meaning == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Word and meaning are required"})
		return
	}

	if err := models.Update(id, userID.(int), input.Word, input.Meaning, input.Example); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update flashcard: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Flashcard updated"})
}

func getFlashcardByID(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	card, err := models.GetByID(id, userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Flashcard not found"})
		return
	}

	c.JSON(http.StatusOK, card)
}

func getDueFlashcards(c *gin.Context) {
	userID, _ := c.Get("user_id")
	cards, err := models.GetDueFlashcards(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to read due flashcards: %v", err)})
		return
	}
	c.JSON(http.StatusOK, cards)
}

func reviewFlashcard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input struct {
		Quality int `json:"quality"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if input.Quality < 0 || input.Quality > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quality must be between 0 and 5"})
		return
	}

	if err := models.UpdateAfterReview(id, userID.(int), input.Quality); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update review: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Flashcard review updated"})
}