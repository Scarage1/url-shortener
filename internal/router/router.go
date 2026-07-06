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

	// --- Services ---

	orgService := service.NewOrgService(db)
	quotaService := service.NewQuotaService(db, redisClient)

	userRepo :=
		repository.NewUserRepository(db)

	authService :=
		service.NewAuthService(
			userRepo,
			orgService,
			cfg.JWTSecret,
			db,
		)

	authHandler :=
		handler.NewAuthHandler(
			authService,
		)

	planHandler := handler.NewPlanHandler(quotaService)

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

	// --- Routes ---

	api := r.Group("/api/v1")

	api.Use(
		middleware.RateLimiter(redisClient),
	)

	// Auth routes (public)
	auth := api.Group("/auth")

	auth.POST(
		"/register",
		middleware.SignupLimiter(redisClient),
		authHandler.Register,
	)

	auth.POST(
		"/login",
		authHandler.Login,
	)

	// Protected routes (require JWT + org membership)
	protected := api.Group("")

	protected.Use(
		middleware.AuthMiddleware(cfg.JWTSecret, orgService, db),
		middleware.PlanRateLimiter(redisClient),
	)

	// Read-only (no email verification needed)
	protected.GET("/plan", planHandler.GetPlan)
	protected.GET("/usage", planHandler.GetUsage)
	protected.GET("/stats/:code", urlHandler.GetStats)
	protected.GET("/links", urlHandler.GetUserLinks)

	// Write operations (require verified email)
	verified := protected.Group("")
	verified.Use(middleware.EmailVerified())

	verified.POST("/shorten", urlHandler.ShortenURL)
	verified.DELETE("/links/:code", urlHandler.DeleteURL)
	verified.GET("/export", urlHandler.ExportLinks)
	verified.POST("/import", urlHandler.ImportLinks)

	// Public redirect (generous rate limit)
	r.Use(middleware.PublicRateLimiter(redisClient))
	r.GET("/:code", urlHandler.RedirectURL)

	return r, urlService
}
