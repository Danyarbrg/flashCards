package models

import (
	"log"
	"time"

	"github.com/Danyarbrg/flashCards/internal/db"
)


type Flashcard struct {
	ID			int		`json:"id"`
	Word	 	string	`json:"word"`
	Meaning		string	`json:"meaning"`
	Example		string	`json:"example"`
	NextReview  time.Time  `json:"next_review"`
    Interval    int     `json:"interval"`
    Repetitions int     `json:"repetitions"`
}

// Flashcard save card into DB.
func (f *Flashcard) Save() error {
    if f.Interval == 0 {
        f.Interval = 1
    }

    now := time.Now().Format("2006-01-02 15:04:05")

    query := `
    INSERT INTO flashcards (word, meaning, example, next_review, interval, repetitions) 
    VALUES (?, ?, ?, ?, ?, ?)`
    
    result, err := db.DB.Exec(query, f.Word, f.Meaning, f.Example, now, f.Interval, f.Repetitions)
    if err != nil {
        log.Printf("Save error: %v", err)
        return err
    }

    lastID, err := result.LastInsertId()
    if err == nil {
        f.ID = int(lastID)
    }
    f.NextReview, _ = time.Parse("2006-01-02 15:04:05", now)

    return nil
}

// GetALL return all cards from DB.
func GetAll() ([]Flashcard, error) {
    rows, err := db.DB.Query("SELECT id, word, meaning, example, next_review, interval, repetitions FROM flashcards")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var cards []Flashcard

    for rows.Next() {
        var f Flashcard
        if err := rows.Scan(&f.ID, &f.Word, &f.Meaning, &f.Example, &f.NextReview, &f.Interval, &f.Repetitions); err != nil {
            log.Println("Error in scan:", err)
        }
        cards = append(cards, f)
    }
    return cards, nil
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
	row := db.DB.QueryRow("SELECT id, word, meaning, example FROM flashcards WHERE id = ?", id)

	var card Flashcard
	err := row.Scan(&card.ID, &card.Word, &card.Meaning, &card.Example)
	return card, err
}

// GetDueFlashcards returns cards for todays repeating.
func GetDueFlashcards() ([]Flashcard, error) {
    query := `SELECT id, word, meaning, example, next_review, interval, repetitions FROM flashcards WHERE next_review <= datetime('now')`
    rows, err := db.DB.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var cards []Flashcard
    for rows.Next() {
        var f Flashcard
        if err := rows.Scan(&f.ID, &f.Word, &f.Meaning, &f.Example, &f.NextReview, &f.Interval, &f.Repetitions); err != nil {
            return nil, err
        }
        cards = append(cards, f)
    }
    return cards, nil
}

// UpdateAfterReview updates rows depending of the success of repetition.
func UpdateAfterReview(id int, success bool) error {
    var query string
    if success {
        query = `
            UPDATE flashcards 
            SET 
                repetitions = repetitions + 1,
                interval = interval * 2,
                next_review = datetime('now', '+' || interval || ' days')
            WHERE id = ?`
    } else {
        query = `
            UPDATE flashcards 
            SET 
                repetitions = 0,
                interval = 1,
                next_review = datetime('now', '+1 day')
            WHERE id = ?`
    }
    _, err := db.DB.Exec(query, id)
    return err
}
