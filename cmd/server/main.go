package main

import (
	"github.com/Scarage1/url-shortener/internal/cache"
	"github.com/Scarage1/url-shortener/internal/config"
	"github.com/Scarage1/url-shortener/internal/database"
	"github.com/Scarage1/url-shortener/internal/logger"
	"github.com/Scarage1/url-shortener/internal/router"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	defer logger.Log.Sync()

	cfg := config.LoadConfig()

	db := database.Connect(cfg)

	redisClient := cache.ConnectRedis(
		cfg.RedisURL,
	)

	r := router.SetupRouter(
		db,
		redisClient,
	)

	logger.Log.Info(
		"Server started on port",

		zap.String(
			"port",
			cfg.Port,
		),
	)
	r.Run(":" + cfg.Port)
}
