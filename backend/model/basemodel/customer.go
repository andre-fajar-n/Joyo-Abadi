package basemodel

import (
	"database/sql"
	"time"
)

type Customer struct {
	CustomerID uint `gorm:"primaryKey"`
	Fullname   string
	Birthday   sql.NullTime
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
