package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Port string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisURL string
}

func LoadConfig() Config {

	err := godotenv.Load()

	if err != nil {

		log.Println(
			"Failed to load .env:",
			err,
		)
	}

	viper.AutomaticEnv()

	viper.SetDefault(
		"PORT",
		"8080",
	)

	return Config{

		Port: viper.GetString("PORT"),

		DBHost: viper.GetString("DB_HOST"),

		DBPort: viper.GetString("DB_PORT"),

		DBUser: viper.GetString("DB_USER"),

		DBPassword: viper.GetString("DB_PASSWORD"),

		DBName: viper.GetString("DB_NAME"),

		RedisURL: viper.GetString("REDIS_URL"),
	}
}
