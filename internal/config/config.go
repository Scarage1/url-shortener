package config

import (
	"fmt"
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

	RedisURL  string
	JWTSecret string
	BaseURL   string
}

var knownWeakSecrets = map[string]bool{
	"":           true,
	"secret-key": true,
	"change-me-to-a-long-random-secret-in-production": true,
	"changeme": true,
	"secret":   true,
}

func LoadConfig() (Config, error) {

	err := godotenv.Load()

	if err != nil {

		log.Println(
			"No .env file loaded (this is expected in production):",
			err,
		)
	}

	viper.AutomaticEnv()

	viper.SetDefault(
		"PORT",
		"8080",
	)

	cfg := Config{

		Port: viper.GetString("PORT"),

		DBHost: viper.GetString("DB_HOST"),

		DBPort: viper.GetString("DB_PORT"),

		DBUser: viper.GetString("DB_USER"),

		DBPassword: viper.GetString("DB_PASSWORD"),

		DBName: viper.GetString("DB_NAME"),

		RedisURL: viper.GetString("REDIS_URL"),

		JWTSecret: viper.GetString("JWT_SECRET"),

		BaseURL: viper.GetString("BASE_URL"),
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validate(cfg Config) error {

	if knownWeakSecrets[cfg.JWTSecret] {
		return fmt.Errorf(
			"JWT_SECRET is empty or a known placeholder value; set a long, random secret via the environment before starting the server",
		)
	}

	if len(cfg.JWTSecret) < 32 {
		return fmt.Errorf(
			"JWT_SECRET is too short (%d chars); use at least 32 random characters", len(cfg.JWTSecret),
		)
	}

	return nil
}
