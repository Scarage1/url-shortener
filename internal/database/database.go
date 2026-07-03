package database

import (
	"fmt"
	"log"

	"github.com/Scarage1/url-shortener/internal/config"
	"github.com/Scarage1/url-shortener/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to PostgreSQL")

	err = db.AutoMigrate(&model.URL{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("Database migrated")

	return db
}
