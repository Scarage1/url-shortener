package router

import (
	"github.com/Scarage1/url-shortener/internal/handler"
	"github.com/Scarage1/url-shortener/internal/repository"
    "github.com/Scarage1/url-shortener/internal/service"

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

	urlRepo := repository.NewURLRepository(db)

	urlService := service.NewURLService(urlRepo)

	urlHandler := handler.NewURLHandler(urlService)

	api := r.Group("/api/v1")

	{
		api.POST("/shorten", urlHandler.ShortenURL)
		api.GET("/stats/:code",urlHandler.GetStats)

	}

	r.GET("/:code", urlHandler.RedirectURL)
	

	return r
}
