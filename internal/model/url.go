package model

import (
	"gorm.io/gorm"
	"time"
)

type URL struct {
	gorm.Model

	OrganizationID uint `gorm:"not null;index"` // ownership + billing + quota
	CreatedBy      uint `gorm:"not null"`        // user who created it (audit trail)

	ShortCode   string `gorm:"uniqueIndex;not null"`
	OriginalURL string

	ClickCount   int
	LastAccessed *time.Time
	Rules        []RoutingRule `gorm:"constraint:OnDelete:CASCADE;"`
}
