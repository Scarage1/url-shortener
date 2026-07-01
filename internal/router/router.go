package router

import (
	"github.com/Scarage1/url-shortener/internal/handler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
		})
	})

	urlHandler := handler.NewURLHandler(db)

	api := r.Group("/api/v1")

	{
		api.POST("/shorten", urlHandler.ShortenURL)

	}

	

	return r
}
