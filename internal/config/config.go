package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	AppName     string
	Port        string
	DatabaseURL string
	RedisURL    string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Failed to load .env: %v\n", err)
	}
	viper.AutomaticEnv()

	cfg := Config{
		AppName:     viper.GetString("APP_NAME"),
		Port:        viper.GetString("PORT"),
		DatabaseURL: viper.GetString("DATABASE_URL"),
		RedisURL:    viper.GetString("REDIS_URL"),
	}
	return cfg
}
