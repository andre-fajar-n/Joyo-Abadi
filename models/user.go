package models

import "gorm.io/gorm"

// User represents a user in the system
type User struct {
	gorm.Model
	Email    string `json:"email" gorm:"unique;not null"`
	Password string `json:"password" gorm:"not null"`
}
