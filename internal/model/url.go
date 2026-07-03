package model

import (
	"time"
	"gorm.io/gorm"
)

type URL struct {
	gorm.Model

	ShortCode string `gorm:"uniqueIndex;not null"`
	OriginalURL string

	ClickCount int
	LastAccessed *time.Time
}