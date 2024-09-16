package models

import "time"

type User struct {
    UserID    string    `json:"user_id"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type BalanceResponse struct {
    Email   string  `json:"email"`
    Balance float64 `json:"balance"`
}