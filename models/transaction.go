package models

import "gorm.io/gorm"

// Transaction represents a sale or purchase transaction
type Transaction struct {
	gorm.Model
	ProductID uint    `json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductID"`
	Quantity  int     `json:"quantity"`
	Total     float64 `json:"total"`
	Type      string  `json:"type"` // "purchase" or "sale"
}
