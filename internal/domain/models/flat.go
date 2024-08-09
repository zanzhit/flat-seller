package models

import "time"

type Flat struct {
	ID         int       `json:"id" db:"id"`
	HouseID    int       `json:"house_id" db:"house_id"`
	Price      int       `json:"price" db:"price"`
	Rooms      int       `json:"rooms" db:"rooms"`
	FlatNumber int       `json:"flat_number" db:"flat_number"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at,omitempty" db:"updated_at"`
}
