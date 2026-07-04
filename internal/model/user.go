package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	UserID uint

	Email string `gorm:"uniqueIndex;not null"`

	PasswordHash string `gorm:"not null"`

	URLs []URL
}
