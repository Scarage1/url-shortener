package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// SignupLimiter rate-limits account creation per IP.
// Limits: 3 per day AND 10 per month per IP address.
// Apply only to POST /auth/register.
func SignupLimiter(redisClient *redis.Client) gin.HandlerFunc {

	return func(c *gin.Context) {

		ctx := context.Background()
		ip := c.ClientIP()

		dailyKey := fmt.Sprintf("signup_ip_daily:%s:%s", ip, time.Now().Format("2006-01-02"))
		monthlyKey := fmt.Sprintf("signup_ip_monthly:%s:%s", ip, time.Now().Format("2006-01"))

		// Check daily limit (3/day)
		dailyCount, err := redisClient.Incr(ctx, dailyKey).Result()
		if err == nil && dailyCount == 1 {
			redisClient.Expire(ctx, dailyKey, 24*time.Hour)
		}

		if dailyCount > 3 {
			c.JSON(
				http.StatusTooManyRequests,
				gin.H{
					"error": "too many registrations from this IP today",
				},
			)
			c.Abort()
			return
		}

		// Check monthly limit (10/month)
		monthlyCount, err := redisClient.Incr(ctx, monthlyKey).Result()
		if err == nil && monthlyCount == 1 {
			redisClient.Expire(ctx, monthlyKey, 35*24*time.Hour)
		}

		if monthlyCount > 10 {
			c.JSON(
				http.StatusTooManyRequests,
				gin.H{
					"error": "too many registrations from this IP this month",
				},
			)
			c.Abort()
			return
		}

		c.Next()
	}
}
