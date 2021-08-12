package basemodel

import (
	"time"
)

type Account struct {
	AccountID uint `gorm:"primaryKey;autoIncrement"`
	Role      string
	Username  string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
	Customer  Customer `gorm:"foreignKey:CustomerID;references:AccountID;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`
}

type AccountRole struct {
	AccountRole string  `gorm:"primaryKey"`
	Account     Account `gorm:"foreignKey:Role;references:AccountRole;constraint:OnUpdate:RESTRICT,OnDelete:RESTRICT;"`
}
