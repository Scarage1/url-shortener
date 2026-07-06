package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Scarage1/url-shortener/internal/model"
	"gorm.io/gorm"
)

// SeedPlans inserts the three pricing tiers if they don't already exist.
// Idempotent — safe to run on every startup.
func SeedPlans(db *gorm.DB) error {

	plans := []model.Plan{
		{
			Name:             model.PlanFree,
			DisplayName:      "Free",
			MaxLinks:         100,
			MaxRedirects:     25_000,
			MaxAPICalls:      1_000,
			MaxDomains:       0,
			MaxGeoRules:      3,
			MaxPasswordLinks: 10,
			MaxScheduleLinks: 10,
			MaxMembers:       1,
			RateLimit:        100,
			PriceMonthly:     0,
			PriceYearly:      0,
		},
		{
			Name:             model.PlanPro,
			DisplayName:      "Pro",
			MaxLinks:         5_000,
			MaxRedirects:     1_000_000,
			MaxAPICalls:      100_000,
			MaxDomains:       3,
			MaxGeoRules:      500,
			MaxPasswordLinks: -1, // unlimited
			MaxScheduleLinks: -1,
			MaxMembers:       3,
			RateLimit:        500,
			PriceMonthly:     499,
			PriceYearly:      4_900,
		},
		{
			Name:             model.PlanBusiness,
			DisplayName:      "Business",
			MaxLinks:         50_000,
			MaxRedirects:     10_000_000,
			MaxAPICalls:      5_000_000,
			MaxDomains:       20,
			MaxGeoRules:      -1,
			MaxPasswordLinks: -1,
			MaxScheduleLinks: -1,
			MaxMembers:       10,
			RateLimit:        2_000,
			PriceMonthly:     1_499,
			PriceYearly:      14_900,
		},
	}

	for _, plan := range plans {

		var existing model.Plan

		result := db.Where("name = ?", plan.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&plan).Error; err != nil {
				return fmt.Errorf("seed plan %s: %w", plan.Name, err)
			}
			log.Printf("seeded plan: %s", plan.Name)
		}
	}

	return nil
}

// MigrateExistingUsers creates a default workspace, org membership, and
// free subscription for every user that doesn't yet belong to an organization.
// Idempotent — safe to run on every startup.
func MigrateExistingUsers(db *gorm.DB) error {

	// Find the free plan ID
	var freePlan model.Plan
	if err := db.Where("name = ?", model.PlanFree).First(&freePlan).Error; err != nil {
		return fmt.Errorf("free plan not found (run SeedPlans first): %w", err)
	}

	// Find all users without an org membership
	var users []model.User

	err := db.Where(
		"id NOT IN (SELECT user_id FROM organization_members WHERE deleted_at IS NULL)",
	).Find(&users).Error

	if err != nil {
		return fmt.Errorf("query orphan users: %w", err)
	}

	for _, user := range users {

		slug := fmt.Sprintf("user_%d", user.ID)
		name := emailToWorkspaceName(user.Email)

		err := db.Transaction(func(tx *gorm.DB) error {

			org := model.Organization{
				Name:    name,
				Slug:    slug,
				OwnerID: user.ID,
			}

			if err := tx.Create(&org).Error; err != nil {
				return err
			}

			member := model.OrganizationMember{
				UserID:         user.ID,
				OrganizationID: org.ID,
				Role:           model.RoleOwner,
			}

			if err := tx.Create(&member).Error; err != nil {
				return err
			}

			sub := model.Subscription{
				OrganizationID:     org.ID,
				PlanID:             freePlan.ID,
				Status:             model.SubscriptionActive,
				CurrentPeriodStart: time.Now(),
				CurrentPeriodEnd:   time.Now().AddDate(100, 0, 0), // free = no expiry
			}

			if err := tx.Create(&sub).Error; err != nil {
				return err
			}

			// Reassign any existing URLs from user_id to the new org
			if err := tx.Model(&model.URL{}).
				Where("user_id = ?", user.ID).
				Update("user_id", org.ID).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.Printf("migrate user %d: %v", user.ID, err)
			continue
		}

		log.Printf("created workspace '%s' for user %d", name, user.ID)
	}

	return nil
}

// emailToWorkspaceName creates a friendly workspace name from an email.
// "shivam@example.com" → "shivam's Workspace"
func emailToWorkspaceName(email string) string {

	parts := strings.Split(email, "@")
	if len(parts) == 0 || parts[0] == "" {
		return "My Workspace"
	}

	return parts[0] + "'s Workspace"
}
