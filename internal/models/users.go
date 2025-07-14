package models

import (
	"fmt"
	"log"

	"github.com/Danyarbrg/flashCards/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

// RegisterUser creating new users with pash password.
func RegisterUser(email, password string) (User, error) {
	var user User
	if email == "" || password == "" {
		return user, fmt.Errorf("email and password are required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return user, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `INSERT INTO users (email, password_hash) VALUES (?, ?)`
	result, err := db.DB.Exec(query, email, hashedPassword)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		return user, fmt.Errorf("failed to register user: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v", err)
		return user, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	user.ID = int(lastID)
	user.Email = email
	user.PasswordHash = string(hashedPassword)
	return user, nil
}

// AuthenticateUser checks your email and passwd and return user in success.
func AuthenticateUser(email, password string) (User, error) {
	var user User
	query := `SELECT id, email, password_hash FROM users WHERE email = ?`
	err := db.DB.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		log.Printf("Failed to find user: %v", err)
		return user, fmt.Errorf("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Printf("Invalid password: %v", err)
		return user, fmt.Errorf("invalid email or password")
	}

	return user, nil
}