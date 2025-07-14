package models
import (
	"log"

	"github.com/Danyarbrg/flashCards/internal/db"
)


type Flashcard struct {
	ID		int		`json:"id"`
	Word 	string	`json:"word"`
	Meaning	string	`json:"meaning"`
	Example	string	`json:"example"`
}

// Flashcard save card into DB.
func (f *Flashcard) Save() error {
	query := `INSERT INTO flashcards (word, meaning, example) VALUES (?, ?, ?)`
	result, err := db.DB.Exec(query, f.Word, f.Meaning, f.Example)
	if err != nil {
		return err
	}
	lastID, err := result.LastInsertId()
	if err == nil {
		f.ID = int(lastID)
	}
	return  nil
}

// GetALL return all cards from DB.
func GetAll() ([]Flashcard, error) {
	rows, err := db.DB.Query("SELECT id, word, meaning, example FROM flashcards")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []Flashcard

	for rows.Next() {
		var f Flashcard
		if err := rows.Scan(&f.ID, &f.Word, &f.Meaning, &f.Example); err != nil {
			log.Println("Error in scan:", err)
		}
		cards = append(cards, f)
	}
	return cards, nil
}

// Delete card from DB.
func Delete(id int) error {
	query := `DELETE FROM slahcards WHERE id = ?`
	_, err := db.DB.Exec(query, id)
	return err
}