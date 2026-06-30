package main

import ("fmt"
		"github.com/Scarage1/url-shortener/internal/config"
		"github.com/Scarage1/url-shortener/internal/database"
		"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	db := database.Connect(cfg)
	_=db // we will use it nxt phase

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"application": cfg.AppName,
			"status": "running",
		})
	})

	fmt.Println("Server started on port", cfg.Port)
	router.Run(":" + cfg.Port)
}
