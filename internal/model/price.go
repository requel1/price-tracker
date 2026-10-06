package model

import "time"

type Price struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	ParsedAt  time.Time `json:"parsed_at"`
}
