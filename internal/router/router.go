package router

import (
	"github.com/Scarage1/url-shortener/internal/config"
	"github.com/Scarage1/url-shortener/internal/geo"
	"github.com/Scarage1/url-shortener/internal/handler"
	"github.com/Scarage1/url-shortener/internal/middleware"
	"github.com/Scarage1/url-shortener/internal/repository"
	"github.com/Scarage1/url-shortener/internal/routing"
	"github.com/Scarage1/url-shortener/internal/security"
	"github.com/Scarage1/url-shortener/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func SetupRouter(
	db *gorm.DB,
	redisClient *redis.Client,
	cfg config.Config,
) (*gin.Engine, *service.URLService) {

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
			cfg.JWTSecret,
		)

	authHandler :=
		handler.NewAuthHandler(
			authService,
		)

	urlRepo := repository.NewURLRepository(db)
	urlScanner := security.NewChainScanner(
		security.NewRulesScanner(cfg.BlockedDomains),
		security.NewGoogleSafeBrowsingScanner(cfg.GoogleSafeBrowsingAPIKey),
	)
	urlResolver := routing.NewEngine()

	urlService := service.NewURLService(
		urlRepo,
		redisClient,
		urlScanner,
		urlResolver,
		geo.NewIPAPILocator(),
	)

	urlHandler := handler.NewURLHandler(
		urlService,
		cfg.BaseURL,
	)

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
		middleware.AuthMiddleware(cfg.JWTSecret),
	)

	protected.POST(
		"/shorten",
		urlHandler.ShortenURL,
	)

	protected.GET(
		"/stats/:code",
		urlHandler.GetStats,
	)

	protected.GET(
		"/links",
		urlHandler.GetUserLinks,
	)

	protected.DELETE(
		"/links/:code",
		urlHandler.DeleteURL,
	)

	protected.GET(
		"/export",
		urlHandler.ExportLinks,
	)

	protected.POST(
		"/import",
		urlHandler.ImportLinks,
	)

	r.GET("/:code", urlHandler.RedirectURL)

	return r, urlService
}
