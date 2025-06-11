package models

import "gorm.io/gorm"

// Product represents a product in the system
type Product struct {
	gorm.Model
	Name        string  `json:"name" gorm:"not null"`
	Price       float64 `json:"price" gorm:"not null"`
	UserID      uint    `json:"user_id" gorm:"nullable"`
	User        *User   `json:"user" gorm:"foreignKey:UserID"`
	Description string  `json:"description"`
}
