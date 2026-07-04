package model

import (
	"gorm.io/gorm"
	"time"
)

type URL struct {
	gorm.Model
	UserID uint

	ShortCode   string `gorm:"uniqueIndex;not null"`
	OriginalURL string

	ClickCount   int
	LastAccessed *time.Time
}
