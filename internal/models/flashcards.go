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
    if f.Interval == 0 {
        f.Interval = 1
    }
    if f.EF == 0 {
        f.EF = 2.5
    }

    now := time.Now().UTC()
    query := `
    INSERT INTO flashcards (word, meaning, example, next_review, interval, repetitions, ef) 
    VALUES (?, ?, ?, ?, ?, ?, ?)`

    result, err := db.DB.Exec(query, f.Word, f.Meaning, f.Example, now.Format(timeFormat), f.Interval, f.Repetitions, f.EF)
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

// GetALL return all cards from DB.
func GetAll() ([]Flashcard, error) {
    rows, err := db.DB.Query("SELECT id, word, meaning, example, next_review, interval, repetitions, ef FROM flashcards")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var cards []Flashcard

    for rows.Next() {
        var f Flashcard
        if err := rows.Scan(&f.ID, &f.Word, &f.Meaning, &f.Example, &f.NextReview, &f.Interval, &f.Repetitions, &f.EF); err != nil {
            log.Println("Error in scan:", err)
        }
        cards = append(cards, f)
    }
    return cards, nil
}

// GetPaginated returns flashcards with limit and offset.
func GetPaginated(limit, offset int) ([]Flashcard, error) {
    query := `SELECT id, word, meaning, example, next_review, interval, repetitions, ef 
				FROM flashcards 
				ORDER BY id 
				LIMIT ? OFFSET ?`

    rows, err := db.DB.Query(query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var cards []Flashcard
    for rows.Next() {
        var f Flashcard
        if err := rows.Scan(&f.ID, &f.Word, &f.Meaning, &f.Example, &f.NextReview, &f.Interval, &f.Repetitions, &f.EF); err != nil {
            log.Println("Error in scan:", err)
        }
        cards = append(cards, f)
    }
    return cards, nil
}

// GetTotalCount returns the total number of flashcards in the DB.
func GetTotalCount() (int, error) {
    var count int
    query := `SELECT COUNT(*) FROM flashcards`
    err := db.DB.QueryRow(query).Scan(&count)
    return count, err
}

// Delete card from DB.
func Delete(id int) error {
	query := `DELETE FROM flashcards WHERE id = ?`
	_, err := db.DB.Exec(query, id)
	return err
}

// Update updates word and translation of the card with specified id.
func Update(id int, word, meaning, example string) error {
	query := `UPDATE flashcards SET word = ?, meaning = ?, example = ? WHERE id = ?`
	_, err := db.DB.Exec(query, word, meaning, example, id)
	return err
}

// GetByID getting card by ID.
func GetByID(id int) (Flashcard, error) {
	row := db.DB.QueryRow("SELECT id, word, meaning, example, next_review, interval, repetitions, ef FROM flashcards WHERE id = ?", id)

	var card Flashcard
	err := row.Scan(&card.ID, &card.Word, &card.Meaning, &card.Example, &card.NextReview, &card.Interval, &card.Repetitions, &card.EF)
	return card, err
}

// GetDueFlashcards returns cards for todays repeating.
func GetDueFlashcards() ([]Flashcard, error) {
    query := `SELECT id, word, meaning, example, next_review, interval, repetitions, ef FROM flashcards WHERE next_review <= datetime('now')`
    rows, err := db.DB.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var cards []Flashcard
    for rows.Next() {
        var f Flashcard
        if err := rows.Scan(&f.ID, &f.Word, &f.Meaning, &f.Example, &f.NextReview, &f.Interval, &f.Repetitions, &f.EF); err != nil {
            return nil, err
        }
        cards = append(cards, f)
    }
    return cards, nil
}

// UpdateAfterReview updates rows depending on review quality using SM-2 algorithm.
func UpdateAfterReview(id int, quality int) error {
    if quality < 0 {
        quality = 0
    } else if quality > 5 {
        quality = 5
    }

    card, err := GetByID(id)
    if err != nil {
        return err
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

        ef = ef + (0.1 - float64(5-quality)*(0.08 + float64(5-quality)*0.02))
        if ef < 1.3 {
            ef = 1.3
        }
    } else {
        card.Repetitions = 0
        card.Interval = 1
    }

    nextReview := time.Now().AddDate(0, 0, card.Interval)

    query := `UPDATE flashcards SET repetitions=?, interval=?, ef=?, next_review=? WHERE id=?`
    _, err = db.DB.Exec(query, card.Repetitions, card.Interval, ef, nextReview.Format("2006-01-02 15:04:05"), id)
    if err != nil {
        log.Printf("UpdateAfterReview DB.Exec error: %v", err)
    }
    return err
}

// GetSortedPaginated realizes sorting and filtration.
func GetSortedPaginated(limit, offset int, sortBy string, asc bool) ([]Flashcard, error) {
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

	query := `
	SELECT id, word, meaning, example, next_review, interval, repetitions, ef
	FROM flashcards
	ORDER BY ` + orderBy + ` ` + orderDir + `
	LIMIT ? OFFSET ?`

	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []Flashcard
	for rows.Next() {
		var f Flashcard
		if err := rows.Scan(&f.ID, &f.Word, &f.Meaning, &f.Example, &f.NextReview, &f.Interval, &f.Repetitions, &f.EF); err != nil {
			log.Println("Error scanning flashcard:", err)
			continue
		}
		cards = append(cards, f)
	}
	return cards, nil
}

// ExistsByWord проверяет, есть ли слово в базе (без учёта регистра).
func ExistsByWord(word string) (bool, error) {
    lowerWord := strings.ToLower(word)
    query := `SELECT COUNT(*) FROM flashcards WHERE LOWER(word) = ?`

    var count int
    err := db.DB.QueryRow(query, lowerWord).Scan(&count)
    if err != nil {
        return false, err
    }

    return count > 0, nil
}