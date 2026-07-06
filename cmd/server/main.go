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

	cfg, err := config.LoadConfig()

	if err != nil {
		logger.Log.Fatal(
			"invalid configuration",
			zap.Error(err),
		)
	}

	db, err := database.Connect(cfg)

	if err != nil {
		logger.Log.Fatal(
			"failed to connect to database",
			zap.Error(err),
		)
	}

	logger.Log.Info("connected to postgres and ran migrations")

	redisClient, err := cache.ConnectRedis(
		cfg.RedisURL,
	)

	if err != nil {
		logger.Log.Fatal(
			"failed to connect to redis",
			zap.Error(err),
		)
	}

	logger.Log.Info("connected to redis")

	r, urlService := router.SetupRouter(
		db,
		redisClient,
		cfg,
	)

	flushCtx, cancelFlush := context.WithCancel(context.Background())
	flushTicker := time.NewTicker(30 * time.Second)
	flushDone := make(chan struct{})

	go func() {
		defer close(flushDone)

		for {
			select {
			case <-flushTicker.C:
				if err := urlService.FlushClickCounts(flushCtx); err != nil {
					logger.Log.Error(
						"click count flush failed",
						zap.Error(err),
					)
				}
			case <-flushCtx.Done():
				return
			}
		}
	}()

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

	flushTicker.Stop()
	cancelFlush()
	<-flushDone

	if err := urlService.FlushClickCounts(context.Background()); err != nil {
		logger.Log.Error(
			"final click count flush failed",
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
