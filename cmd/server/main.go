package main

import ("fmt"
		"github.com/Scarage1/url-shortener/internal/config"
		"github.com/Scarage1/url-shortener/internal/database"
		"github.com/Scarage1/url-shortener/internal/router"
)

func main() {
	cfg := config.LoadConfig()
	db := database.Connect(cfg)
	r := router.SetupRouter(db)

	

	fmt.Println("Server started on port", cfg.Port)
	r.Run(":" + cfg.Port)
}
