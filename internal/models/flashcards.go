package models

import (
	"log"
	"time"
	"strings"
	"errors"
	"fmt"

	"github.com/Danyarbrg/flashCards/internal/db"
)


type Flashcard struct {
	ID			int			`json:"id"`
	UserID		int			`json:"user_id"`
	Word	 	string		`json:"word"`
	Meaning		string		`json:"meaning"`
	Example		string		`json:"example"`
	NextReview  time.Time   `json:"next_review"`
    Interval    int     	`json:"interval"`
    Repetitions int     	`json:"repetitions"`
    EF			float64		`json:"ef"`
}

const timeFormat = "2006-01-02 15:04:05"

// Flashcard save card into DB.
func (f *Flashcard) Save() error {
    if f.Word == "" || f.Meaning == "" {
        return errors.New("word and meaning are required")
    }
	if f.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}
    if f.Interval == 0 {
        f.Interval = 1
    }
    if f.EF == 0 {
        f.EF = 2.5
    }

    now := time.Now().UTC()
	query := `
	INSERT INTO flashcards (user_id, word, meaning, example, next_review, interval, repetitions, ef) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

    result, err := db.DB.Exec(query, f.UserID, f.Word, f.Meaning, f.Example, now.Format(timeFormat), f.Interval, f.Repetitions, f.EF)
	if err != nil {
		log.Printf("Failed to save flashcard: %v", err)
		return fmt.Errorf("failed to save flashcard: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v", err)
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}
	f.ID = int(lastID)
	f.NextReview = now

	return nil
}

func GetAll(userID int) ([]Flashcard, error) {
	query := `SELECT id, user_id, word, meaning, example, next_review, interval, repetitions, ef 
			FROM flashcards 
			WHERE user_id = ?`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query flashcards: %w", err)
	}
	defer rows.Close()

	var cards []Flashcard
	for rows.Next() {
		var f Flashcard
		var nextReviewStr string
		if err := rows.Scan(&f.ID, &f.UserID, &f.Word, &f.Meaning, &f.Example, &nextReviewStr, &f.Interval, &f.Repetitions, &f.EF); err != nil {
			log.Printf("Error scanning flashcard: %v", err)
			return nil, fmt.Errorf("failed to scan flashcard: %w", err)
		}
		f.NextReview, err = time.Parse(timeFormat, nextReviewStr)
		if err != nil {
			log.Printf("Error parsing next_review: %v", err)
			return nil, fmt.Errorf("failed to parse next_review: %w", err)
		}
		cards = append(cards, f)
	}
	return cards, nil
}

func GetPaginated(userID, limit, offset int) ([]Flashcard, error) {
	query := `SELECT id, user_id, word, meaning, example, next_review, interval, repetitions, ef 
			FROM flashcards 
			WHERE user_id = ?
			ORDER BY id 
			LIMIT ? OFFSET ?`

	rows, err := db.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query flashcards: %w", err)
	}
	defer rows.Close()

	var cards []Flashcard
	for rows.Next() {
		var f Flashcard
		var nextReviewStr string
		if err := rows.Scan(&f.ID, &f.UserID, &f.Word, &f.Meaning, &f.Example, &nextReviewStr, &f.Interval, &f.Repetitions, &f.EF); err != nil {
			log.Printf("Error scanning flashcard: %v", err)
			return nil, fmt.Errorf("failed to scan flashcard: %w", err)
		}
		f.NextReview, err = time.Parse(timeFormat, nextReviewStr)
		if err != nil {
			log.Printf("Error parsing next_review: %v", err)
			return nil, fmt.Errorf("failed to parse next_review: %w", err)
		}
		cards = append(cards, f)
	}
	return cards, nil
}

func GetTotalCount(userID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM flashcards WHERE user_id = ?`
	err := db.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}
	return count, nil
}

func Delete(id, userID int) error {
	query := `DELETE FROM flashcards WHERE id = ? AND user_id = ?`
	_, err := db.DB.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete flashcard: %w", err)
	}
	return nil
}

func Update(id, userID int, word, meaning, example string) error {
	query := `UPDATE flashcards SET word = ?, meaning = ?, example = ? WHERE id = ? AND user_id = ?`
	_, err := db.DB.Exec(query, word, meaning, example, id, userID)
	if err != nil {
		return fmt.Errorf("failed to update flashcard: %w", err)
	}
	return nil
}

func GetByID(id, userID int) (Flashcard, error) {
	query := `SELECT id, user_id, word, meaning, example, next_review, interval, repetitions, ef 
			FROM flashcards 
			WHERE id = ? AND user_id = ?`
	row := db.DB.QueryRow(query, id, userID)

	var card Flashcard
	var nextReviewStr string
	err := row.Scan(&card.ID, &card.UserID, &card.Word, &card.Meaning, &card.Example, &nextReviewStr, &card.Interval, &card.Repetitions, &card.EF)
	if err != nil {
		return card, fmt.Errorf("failed to get flashcard: %w", err)
	}
	card.NextReview, err = time.Parse(timeFormat, nextReviewStr)
	if err != nil {
		return card, fmt.Errorf("failed to parse next_review: %w", err)
	}
	return card, nil
}

func GetDueFlashcards(userID int) ([]Flashcard, error) {
	now := time.Now().UTC()
	query := `SELECT id, user_id, word, meaning, example, next_review, interval, repetitions, ef 
			FROM flashcards 
			WHERE user_id = ? AND next_review <= ?`
	rows, err := db.DB.Query(query, userID, now.Format(timeFormat))
	if err != nil {
		return nil, fmt.Errorf("failed to query due flashcards: %w", err)
	}
	defer rows.Close()

	var cards []Flashcard
	for rows.Next() {
		var f Flashcard
		var nextReviewStr string
		if err := rows.Scan(&f.ID, &f.UserID, &f.Word, &f.Meaning, &f.Example, &nextReviewStr, &f.Interval, &f.Repetitions, &f.EF); err != nil {
			log.Printf("Error scanning flashcard: %v", err)
			return nil, fmt.Errorf("failed to scan flashcard: %w", err)
		}
		f.NextReview, err = time.Parse(timeFormat, nextReviewStr)
		if err != nil {
			log.Printf("Error parsing next_review: %v", err)
			return nil, fmt.Errorf("failed to parse next_review: %w", err)
		}
		cards = append(cards, f)
	}
	return cards, nil
}

func UpdateAfterReview(id, userID, quality int) error {
	if quality < 0 {
		quality = 0
	} else if quality > 5 {
		quality = 5
	}

	card, err := GetByID(id, userID)
	if err != nil {
		return fmt.Errorf("failed to get flashcard: %w", err)
	}

	ef := card.EF
	if ef == 0 {
		ef = 2.5
	}

	if quality >= 3 {
		if card.Repetitions == 0 {
			card.Interval = 1
		} else if card.Repetitions == 1 {
			card.Interval = 6
		} else {
			card.Interval = int(float64(card.Interval) * ef)
		}
		card.Repetitions++
		ef = ef + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
		if ef < 1.3 {
			ef = 1.3
		}
	} else {
		card.Repetitions = 0
		card.Interval = 1
	}

	nextReview := time.Now().UTC().AddDate(0, 0, card.Interval)
	query := `UPDATE flashcards SET repetitions = ?, interval = ?, ef = ?, next_review = ? 
			WHERE id = ? AND user_id = ?`
	_, err = db.DB.Exec(query, card.Repetitions, card.Interval, ef, nextReview.Format(timeFormat), id, userID)
	if err != nil {
		log.Printf("UpdateAfterReview error: %v", err)
		return fmt.Errorf("failed to update flashcard: %w", err)
	}
	return nil
}

func GetSortedPaginated(userID, limit, offset int, sortBy string, asc bool) ([]Flashcard, error) {
	validSortFields := map[string]string{
		"created":     "id",
		"repetitions": "repetitions",
		"ef":          "ef",
		"next_review": "next_review",
	}

	orderBy, ok := validSortFields[sortBy]
	if !ok {
		orderBy = "id"
	}

	orderDir := "ASC"
	if !asc {
		orderDir = "DESC"
	}

	query := fmt.Sprintf(`
	SELECT id, user_id, word, meaning, example, next_review, interval, repetitions, ef
	FROM flashcards
	WHERE user_id = ?
	ORDER BY %s %s
	LIMIT ? OFFSET ?`, orderBy, orderDir)

	rows, err := db.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query flashcards: %w", err)
	}
	defer rows.Close()

	var cards []Flashcard
	for rows.Next() {
		var f Flashcard
		var nextReviewStr string
		if err := rows.Scan(&f.ID, &f.UserID, &f.Word, &f.Meaning, &f.Example, &nextReviewStr, &f.Interval, &f.Repetitions, &f.EF); err != nil {
			log.Printf("Error scanning flashcard: %v", err)
			return nil, fmt.Errorf("failed to scan flashcard: %w", err)
		}
		f.NextReview, err = time.Parse(timeFormat, nextReviewStr)
		if err != nil {
			log.Printf("Error parsing next_review: %v", err)
			return nil, fmt.Errorf("failed to parse next_review: %w", err)
		}
		cards = append(cards, f)
	}
	return cards, nil
}

func ExistsByWord(userID int, word string) (bool, error) {
	lowerWord := strings.ToLower(word)
	query := `SELECT COUNT(*) FROM flashcards WHERE user_id = ? AND LOWER(word) = ?`

	var count int
	err := db.DB.QueryRow(query, userID, lowerWord).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check word existence: %w", err)
	}

	return count > 0, nil
}