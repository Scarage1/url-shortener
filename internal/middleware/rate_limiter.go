package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimiter(
	redisClient *redis.Client,
) gin.HandlerFunc {

	return func(c *gin.Context) {

		ctx := context.Background()

		ip := c.ClientIP()

		key := "rate_limit:" + ip

		count, err :=
			redisClient.Incr(
				ctx,
				key,
			).Result()

		if err != nil {

			c.Next()
			return
		}

		if count == 1 {

			redisClient.Expire(
				ctx,
				key,
				time.Minute,
			)
		}

		if count > 10 {

			c.JSON(
				http.StatusTooManyRequests,
				gin.H{
					"error": "too many requests",
				},
			)

			c.Abort()
			return
		}

		c.Next()
	}
}
