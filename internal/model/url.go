package model

import (
	"gorm.io/gorm"
)

type URL struct {
	gorm.Model

	ShortCode string `gorm:"uniqueIndex;not null"`
	OriginalURL string
}