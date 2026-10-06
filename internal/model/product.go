package model

import "time"

type Product struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	Name      string    `json:"name"`
	Site      string    `json:"site"`
	CreatedAt time.Time `json:"created_at"`
}
