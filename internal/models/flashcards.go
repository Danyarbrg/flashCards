package models

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Danyarbrg/flashCards/internal/db"
)

const timeFormat = "2006-01-02T15:04:05Z"

type Flashcard struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Word        string    `json:"word"`
	Meaning     string    `json:"meaning"`
	Example     string    `json:"example"`
	Tags        string    `json:"tags"`
	NextReview  time.Time `json:"next_review"`
	Interval    int       `json:"interval"`
	Repetitions int       `json:"repetitions"`
	EF          float64   `json:"ef"`
	CreatedAt   time.Time `json:"created_at"`
}

func (f *Flashcard) Save() error {
	if f.Word == "" || f.Meaning == "" {
		return fmt.Errorf("word and meaning are required")
	}
	if f.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}

	now := time.Now().UTC()
	query := `
	INSERT INTO flashcards (user_id, word, meaning, example, tags, next_review, interval, repetitions, ef, created_at) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.DB.Exec(query, f.UserID, f.Word, f.Meaning, f.Example, f.Tags, now.Format(timeFormat), 1, 0, 2.5, now.Format(timeFormat))
	if err != nil {
		return fmt.Errorf("failed to save flashcard: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	f.ID = int(lastID)
	f.NextReview = now
	f.CreatedAt = now
	return nil
}

func Update(id, userID int, word, meaning, example, tags string) error {
	query := `UPDATE flashcards SET word = ?, meaning = ?, example = ?, tags = ? WHERE id = ? AND user_id = ?`
	_, err := db.DB.Exec(query, word, meaning, example, tags, id, userID)
	if err != nil {
		return fmt.Errorf("failed to update flashcard: %w", err)
	}
	return nil
}

func GetSortedPaginated(userID, limit, offset int, sortBy, order, tagFilter string) ([]Flashcard, error) {
	validSortFields := map[string]string{
		"created":     "created_at",
		"repetitions": "repetitions",
		"ef":          "ef",
		"next_review": "next_review",
	}
	orderBy, ok := validSortFields[sortBy]
	if !ok {
		orderBy = "created_at"
	}
	orderDir := "ASC"
	if strings.ToLower(order) == "desc" {
		orderDir = "DESC"
	}

	baseQuery := `SELECT id, user_id, word, meaning, example, tags, next_review, interval, repetitions, ef, created_at FROM flashcards WHERE user_id = ?`
	args := []interface{}{userID}

	if tagFilter != "" {
		baseQuery += " AND LOWER(tags) LIKE ?"
		args = append(args, "%"+strings.ToLower(tagFilter)+"%")
	}

	fullQuery := fmt.Sprintf("%s ORDER BY %s %s LIMIT ? OFFSET ?", baseQuery, orderBy, orderDir)
	args = append(args, limit, offset)

	rows, err := db.DB.Query(fullQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sorted flashcards: %w", err)
	}
	defer rows.Close()

	var cards []Flashcard
	for rows.Next() {
		var f Flashcard
		var nextReviewStr, createdAtStr string
		if err := rows.Scan(&f.ID, &f.UserID, &f.Word, &f.Meaning, &f.Example, &f.Tags, &nextReviewStr, &f.Interval, &f.Repetitions, &f.EF, &createdAtStr); err != nil {
			log.Printf("Error scanning flashcard: %v", err)
			return nil, fmt.Errorf("failed to scan flashcard: %w", err)
		}
		f.NextReview, _ = time.Parse(timeFormat, nextReviewStr)
		f.CreatedAt, _ = time.Parse(timeFormat, createdAtStr)
		cards = append(cards, f)
	}
	return cards, nil
}

func GetByID(id, userID int) (Flashcard, error) {
	query := `SELECT id, user_id, word, meaning, example, tags, next_review, interval, repetitions, ef, created_at 
			FROM flashcards 
			WHERE id = ? AND user_id = ?`
	row := db.DB.QueryRow(query, id, userID)

	var card Flashcard
	var nextReviewStr, createdAtStr string
	err := row.Scan(&card.ID, &card.UserID, &card.Word, &card.Meaning, &card.Example, &card.Tags, &nextReviewStr, &card.Interval, &card.Repetitions, &card.EF, &createdAtStr)
	if err != nil {
		return card, fmt.Errorf("failed to get flashcard: %w", err)
	}
	card.NextReview, _ = time.Parse(timeFormat, nextReviewStr)
	card.CreatedAt, _ = time.Parse(timeFormat, createdAtStr)
	return card, nil
}

func GetDueFlashcards(userID int) ([]Flashcard, error) {
    now := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, 1)
    query := `SELECT id, user_id, word, meaning, example, tags, next_review, interval, repetitions, ef, created_at 
            FROM flashcards 
            WHERE user_id = ? AND next_review < ?`
    rows, err := db.DB.Query(query, userID, now.Format(timeFormat))
    if err != nil {
        return nil, fmt.Errorf("failed to query due flashcards: %w", err)
    }
    defer rows.Close()

    var cards []Flashcard
    for rows.Next() {
        var f Flashcard
        var nextReviewStr, createdAtStr string
        if err := rows.Scan(&f.ID, &f.UserID, &f.Word, &f.Meaning, &f.Example, &f.Tags, &nextReviewStr, &f.Interval, &f.Repetitions, &f.EF, &createdAtStr); err != nil {
            log.Printf("Error scanning flashcard: %v", err)
            return nil, fmt.Errorf("failed to scan flashcard: %w", err)
        }
        f.NextReview, _ = time.Parse(timeFormat, nextReviewStr)
        f.CreatedAt, _ = time.Parse(timeFormat, createdAtStr)
        cards = append(cards, f)
    }
    return cards, nil
}

func UpdateAfterReview(id, userID, quality int) error {
	card, err := GetByID(id, userID)
	if err != nil {
		return fmt.Errorf("failed to get flashcard: %w", err)
	}

	if quality < 3 {
		card.Repetitions = 0
		card.Interval = 1
	} else {
		if card.Repetitions == 0 {
			card.Interval = 1
		} else if card.Repetitions == 1 {
			card.Interval = 6
		} else {
			card.Interval = int(float64(card.Interval) * card.EF)
		}
		card.Repetitions++
		card.EF += (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
		if card.EF < 1.3 {
			card.EF = 1.3
		}
	}

	nextReview := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, card.Interval)

	query := `UPDATE flashcards SET repetitions = ?, interval = ?, ef = ?, next_review = ? 
			WHERE id = ? AND user_id = ?`
	_, err = db.DB.Exec(query, card.Repetitions, card.Interval, card.EF, nextReview.Format(timeFormat), id, userID)
	if err != nil {
		return fmt.Errorf("failed to update flashcard: %w", err)
	}
	return nil
}

func Delete(id, userID int) error {
	query := `DELETE FROM flashcards WHERE id = ? AND user_id = ?`
	_, err := db.DB.Exec(query, id, userID)
	return err
}

func ExistsByWord(userID int, word string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM flashcards WHERE user_id = ? AND LOWER(word) = ?`
	err := db.DB.QueryRow(query, userID, strings.ToLower(word)).Scan(&count)
	return count > 0, err
}

func GetAllTags(userID int) ([]string, error) {
	query := `SELECT tags FROM flashcards WHERE user_id = ? AND tags != ''`
	rows, err := db.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	uniqueTags := make(map[string]bool)

	for rows.Next() {
		var tagsStr string
		if err := rows.Scan(&tagsStr); err != nil {
			return nil, fmt.Errorf("failed to scan tags: %w", err)
		}

		tags := strings.Split(tagsStr, ",")
		for _, tag := range tags {
			trimmedTag := strings.TrimSpace(tag)
			if trimmedTag != "" {
				uniqueTags[trimmedTag] = true
			}
		}
	}

	var result []string
	for tag := range uniqueTags {
		result = append(result, tag)
	}

	return result, nil
}