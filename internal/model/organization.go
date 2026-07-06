package model

import "gorm.io/gorm"

// Organization member role constants.
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// Organization is the billing and ownership unit.
// Every user belongs to at least one org ("Shivam's Workspace").
// URLs, subscriptions, and usage counters are scoped to the org.
type Organization struct {
	gorm.Model

	Name    string `gorm:"not null"`
	Slug    string `gorm:"uniqueIndex;not null"`
	OwnerID uint   `gorm:"not null;index"` // fast ownership check without querying members

	Members []OrganizationMember
	URLs    []URL
}

// OrganizationMember is the join table between users and organizations.
type OrganizationMember struct {
	gorm.Model

	UserID         uint `gorm:"not null;uniqueIndex:idx_org_member"`
	OrganizationID uint `gorm:"not null;uniqueIndex:idx_org_member"`

	Role string `gorm:"not null;default:'member'"` // owner, admin, member
}
