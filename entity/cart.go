package entity

import (
	"time"
)

// Cart Lifecycle
// Cart Created --> Active
// New Items added --> Active
// Items cleared --> Cleared
// Order placed aka purchased --> Completed
type CartStatus string

const (
	CartStatusActive    CartStatus = "active"
	CartStatusCompleted CartStatus = "completed"
	CartStatusCleared   CartStatus = "cleared"
)

type Cart struct {
	ID        string     `json:"id"`
	Version   int        `json:"version"`
	UserID    string     `json:"user_id"`
	Items     []CartItem `json:"items"`
	Status    CartStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CartItem struct {
	ID        string `json:"id"`
	Quantity  int    `json:"quantity"`
	ProductId string `json:"product_id"`
}
