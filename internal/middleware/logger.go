package middleware

import (
	"time"

	"github.com/Scarage1/url-shortener/internal/logger"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

func Logger() gin.HandlerFunc {

	return func(c *gin.Context) {

		start := time.Now()

		c.Next()

		requestID, _ :=
			c.Get(
				"request_id",
			)

		userIDValue, _ := c.Get("user_id")
		orgIDValue, _ := c.Get("org_id")

		logger.Log.Info(
			"request completed",

			zap.Any(
				"request_id",
				requestID,
			),

			zap.String(
				"method",
				c.Request.Method,
			),

			zap.String(
				"path",
				c.Request.URL.Path,
			),

			zap.Int(
				"status",
				c.Writer.Status(),
			),

			zap.Duration(
				"latency",
				time.Since(start),
			),

			zap.String(
				"client_ip",
				c.ClientIP(),
			),

			zap.Any(
				"user_id",
				userIDValue,
			),

			zap.Any(
				"org_id",
				orgIDValue,
			),
		)
	}
}
