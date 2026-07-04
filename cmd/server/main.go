package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		cfg,
	)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// START SERVER FIRST
	go func() {

		logger.Log.Info(
			"server started",
			zap.String(
				"port",
				cfg.Port,
			),
		)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			logger.Log.Fatal(
				"server failed",
				zap.Error(err),
			)
		}
	}()

	// THEN WAIT FOR SHUTDOWN SIGNAL
	quit := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		quit,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-quit

	logger.Log.Info(
		"shutting down server",
	)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {

		logger.Log.Fatal(
			"forced shutdown",
			zap.Error(err),
		)
	}

	if err := redisClient.Close(); err != nil {

		logger.Log.Error(
			"redis close error",
			zap.Error(err),
		)
	}

	logger.Log.Info(
		"server exited",
	)
}
