package main

import ("fmt"
		"github.com/Scarage1/url-shortener/internal/config"
		"github.com/Scarage1/url-shortener/internal/database"
		"github.com/Scarage1/url-shortener/internal/router"
		"github.com/Scarage1/url-shortener/internal/cache"
)

func main() {
	cfg := config.LoadConfig()
	
	db := database.Connect(cfg)

	redisClient := cache.ConnectRedis(
	cfg.RedisURL,
	)

	r := router.SetupRouter(
		db,
		redisClient,
	)

	

	fmt.Println("Server started on port", cfg.Port)
	r.Run(":" + cfg.Port)
}
