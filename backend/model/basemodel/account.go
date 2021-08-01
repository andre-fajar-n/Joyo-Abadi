package basemodel

import (
	"database/sql"

	"gorm.io/gorm"
)

type AccountBaseModel struct {
	gorm.Model
	Username string `gorm:""`
	Password string
	Birtday  sql.NullTime
}
