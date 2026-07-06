package model

import "gorm.io/gorm"

// Plan name constants used in database seeds and lookups.
const (
	PlanFree     = "free"
	PlanPro      = "pro"
	PlanBusiness = "business"
)

// Plan defines a SaaS pricing tier with its feature limits.
// Limits set to -1 mean unlimited.
type Plan struct {
	gorm.Model

	Name        string `gorm:"uniqueIndex;not null"` // "free", "pro", "business"
	DisplayName string `gorm:"not null"`             // "Free", "Pro", "Business"

	// Resource limits
	MaxLinks         int // active links (-1 = unlimited)
	MaxRedirects     int // per month
	MaxAPICalls      int // per month
	MaxDomains       int // custom domains
	MaxGeoRules      int // geo routing destinations
	MaxPasswordLinks int // password-protected links
	MaxScheduleLinks int // scheduled links
	MaxMembers       int // team members per org

	// Rate limiting
	RateLimit int // requests per minute per user

	// Pricing in EUR cents (0 = free)
	PriceMonthly int
	PriceYearly  int
}
