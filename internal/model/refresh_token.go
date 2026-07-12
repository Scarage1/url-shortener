package model

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken stores hashed refresh token sessions for users.
type RefreshToken struct {
	gorm.Model
	UserID    uint       `gorm:"not null;index"`
	TokenHash string     `gorm:"not null;uniqueIndex"`
	ExpiresAt time.Time  `gorm:"not null"`
	RevokedAt *time.Time `gorm:"index"`
}
