package model

import (
	"time"

	"gorm.io/gorm"
)

// Subscription status constants.
const (
	SubscriptionActive    = "active"
	SubscriptionCancelled = "cancelled"
	SubscriptionExpired   = "expired"
)

// Subscription connects an organization to a plan.
// This is the billing record — Stripe webhooks update status here.
// Each org has exactly one active subscription at a time.
type Subscription struct {
	gorm.Model

	OrganizationID uint   `gorm:"uniqueIndex;not null"`
	PlanID         uint   `gorm:"not null"`
	Plan           Plan   `gorm:"constraint:OnDelete:RESTRICT;"`
	Status         string `gorm:"not null;default:'active'"` // active, cancelled, expired

	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
}
