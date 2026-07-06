package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var rateLimitScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
if count == 1 then
	redis.call("EXPIRE", KEYS[1], ARGV[1])
end
return count
`)

const (
	rateLimitWindow = time.Minute
	rateLimitMax    = 10
)

func RateLimiter(
	redisClient *redis.Client,
) gin.HandlerFunc {

	return func(c *gin.Context) {

		ctx := context.Background()

		ip := c.ClientIP()

		key := "rate_limit:" + ip

		count, err := rateLimitScript.Run(
			ctx,
			redisClient,
			[]string{key},
			int(rateLimitWindow.Seconds()),
		).Int()

		if err != nil {

			c.Next()
			return
		}

		if count > rateLimitMax {

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
