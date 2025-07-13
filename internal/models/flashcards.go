package models

type Flashcard struct {
	ID		int		`json:"id"`
	Word 	string	`json:"word"`
	Meaning	string	`json:"meaning"`
	Example	string	`json:"example"`
}