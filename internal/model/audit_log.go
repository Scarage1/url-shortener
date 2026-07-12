package model

import (
	"time"
)

// AuditLog records security-relevant events within an organization.
type AuditLog struct {
	ID             uint   `gorm:"primaryKey"`
	OrganizationID uint   `gorm:"not null;index"`
	ActorID        uint   `gorm:"not null"`
	Action         string `gorm:"not null;index"` // "link.created", "link.deleted", "user.registered", "user.login", etc.
	ResourceType   string `gorm:"not null"`       // "url", "user", "member"
	ResourceID     string `gorm:"not null"`       // short_code, user_id, etc.
	Metadata       string `gorm:"type:text"`      // JSON string
	IPAddress      string
	CreatedAt      time.Time `gorm:"not null;index"`
}
