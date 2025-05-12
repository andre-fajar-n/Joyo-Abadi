package models

import "gorm.io/gorm"

// Product represents a product in the system
type Product struct {
	gorm.Model
	Name     string  `json:"name" gorm:"not null"`
	Price    float64 `json:"price" gorm:"not null"`
	Quantity int     `json:"quantity" gorm:"not null"`
}
