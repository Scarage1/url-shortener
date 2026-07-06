package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/Scarage1/url-shortener/internal/model"
	"gorm.io/gorm"
)

// OrgService handles workspace creation and membership lookups.
type OrgService struct {
	DB *gorm.DB
}

func NewOrgService(db *gorm.DB) *OrgService {
	return &OrgService{DB: db}
}

// GetOrgIDForUser returns the user's primary organization ID.
// Every user has exactly one org (created at registration).
func (s *OrgService) GetOrgIDForUser(userID uint) (uint, error) {

	var member model.OrganizationMember

	err := s.DB.Where(
		"user_id = ?",
		userID,
	).First(&member).Error

	if err != nil {
		return 0, fmt.Errorf("org not found for user %d: %w", userID, err)
	}

	return member.OrganizationID, nil
}

// CreateDefaultOrg creates a personal workspace, owner membership, and
// free subscription for a newly registered user. Must be called inside
// the registration transaction.
func (s *OrgService) CreateDefaultOrg(tx *gorm.DB, user *model.User, freePlanID uint) error {

	slug := fmt.Sprintf("user_%d", user.ID)
	name := emailToWorkspaceName(user.Email)

	org := model.Organization{
		Name:    name,
		Slug:    slug,
		OwnerID: user.ID,
	}

	if err := tx.Create(&org).Error; err != nil {
		return fmt.Errorf("create org: %w", err)
	}

	member := model.OrganizationMember{
		UserID:         user.ID,
		OrganizationID: org.ID,
		Role:           model.RoleOwner,
	}

	if err := tx.Create(&member).Error; err != nil {
		return fmt.Errorf("create membership: %w", err)
	}

	sub := model.Subscription{
		OrganizationID:     org.ID,
		PlanID:             freePlanID,
		Status:             model.SubscriptionActive,
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(100, 0, 0), // free = no expiry
	}

	if err := tx.Create(&sub).Error; err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}

	return nil
}

func emailToWorkspaceName(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 0 || parts[0] == "" {
		return "My Workspace"
	}
	return parts[0] + "'s Workspace"
}
