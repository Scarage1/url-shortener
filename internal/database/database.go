package database

import (
	"fmt"

	"github.com/Scarage1/url-shortener/internal/config"
	"github.com/Scarage1/url-shortener/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.Config) *gorm.DB {

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

		panic(err)
	}

	fmt.Println(
		"Connected to PostgreSQL",
	)

	err = db.AutoMigrate(
		&model.User{},
		&model.URL{},
	)

	if err != nil {

		panic(err)
	}

	fmt.Println(
		"Database migrated",
	)

	return db
}
