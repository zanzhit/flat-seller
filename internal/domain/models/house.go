package models

import "time"

type House struct {
	ID        int       `json:"id" db:"id"`
	Address   string    `json:"address" db:"address"`
	Year      int       `json:"year" db:"year"`
	Developer string    `json:"developer,omitempty" db:"developer"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
}
