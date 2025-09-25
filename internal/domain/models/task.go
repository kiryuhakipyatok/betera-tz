package models

import "github.com/google/uuid"

type Task struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Descpription string    `json:"description"`
	Status       string    `json:"status"`
}
