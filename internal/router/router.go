package router

import (
	"github.com/Scarage1/url-shortener/internal/handler"
	"github.com/Scarage1/url-shortener/internal/middleware"
	"github.com/Scarage1/url-shortener/internal/repository"
	"github.com/Scarage1/url-shortener/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func SetupRouter(
	db *gorm.DB,
	redisClient *redis.Client,
) *gin.Engine {

	r := gin.New()

	r.Use(
		gin.Recovery(),
	)

	r.Use(
		middleware.RequestID(),
		middleware.Logger(),
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	userRepo :=
		repository.NewUserRepository(db)

	authService :=
		service.NewAuthService(
			userRepo,
		)

	authHandler :=
		handler.NewAuthHandler(
			authService,
		)

	urlRepo := repository.NewURLRepository(db)

	urlService := service.NewURLService(
		urlRepo,
		redisClient,
	)

	urlHandler := handler.NewURLHandler(urlService)

	api := r.Group("/api/v1")

	api.Use(
		middleware.RateLimiter(
			redisClient,
		),
	)

	auth :=
		api.Group(
			"/auth",
		)

	auth.POST(
		"/register",
		authHandler.Register,
	)

	auth.POST(
		"/login",
		authHandler.Login,
	)

	protected :=
		api.Group(
			"",
		)

	protected.Use(
		middleware.AuthMiddleware(),
	)

	protected.POST(
		"/shorten",
		urlHandler.ShortenURL,
	)

	protected.GET(
		"/stats/:code",
		urlHandler.GetStats,
	)

	r.GET("/:code", urlHandler.RedirectURL)

	return r
}
