package database

import (
	"fmt"

	"github.com/Scarage1/url-shortener/internal/config"
	"github.com/Scarage1/url-shortener/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.Config) (*gorm.DB, error) {

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{},
	)

	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	err = db.AutoMigrate(
		&model.User{},
		&model.URL{},
		&model.RoutingRule{},
		&model.Plan{},
		&model.Organization{},
		&model.OrganizationMember{},
		&model.Subscription{},
	)

	if err != nil {
		return nil, fmt.Errorf("auto-migrate: %w", err)
	}

	// Seed pricing tiers (idempotent)
	if err := SeedPlans(db); err != nil {
		return nil, fmt.Errorf("seed plans: %w", err)
	}

	// Create default workspaces for any existing users without orgs (idempotent)
	if err := MigrateExistingUsers(db); err != nil {
		return nil, fmt.Errorf("migrate existing users: %w", err)
	}

	return db, nil
}
