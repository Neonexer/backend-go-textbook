package model

import "time"

// Product — товар на маркетплейсе.
type Product struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int       `json:"price"` // в копейках
	SellerID    int       `json:"seller_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User — пользователь маркетплейса.
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // не сериализуем в JSON
	Role         string    `json:"role"` // buyer, seller, admin
	CreatedAt    time.Time `json:"created_at"`
}

// ErrorResponse — ошибка API.
type ErrorResponse struct {
	Error string `json:"error"`
}
