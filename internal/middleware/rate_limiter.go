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

// PublicRateLimiter applies a generous IP-based rate limit for public routes
// (redirects). 300 req/min per IP to avoid blocking legitimate viral traffic
// behind shared NATs (college/corporate networks).
func PublicRateLimiter(redisClient *redis.Client) gin.HandlerFunc {

	const (
		window = time.Minute
		limit  = 300
	)

	return func(c *gin.Context) {

		ctx := context.Background()
		ip := c.ClientIP()
		key := "rate_public:" + ip

		count, err := rateLimitScript.Run(
			ctx,
			redisClient,
			[]string{key},
			int(window.Seconds()),
		).Int()

		if err != nil {
			c.Next()
			return
		}

		if count > limit {
			c.JSON(
				http.StatusTooManyRequests,
				gin.H{"error": "too many requests"},
			)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiter applies an IP-based rate limit for the API group (pre-auth).
// This is a baseline protection — the plan-aware limiter runs after auth.
func RateLimiter(redisClient *redis.Client) gin.HandlerFunc {

	const (
		window = time.Minute
		limit  = 60 // generous pre-auth baseline
	)

	return func(c *gin.Context) {

		ctx := context.Background()
		ip := c.ClientIP()
		key := "rate_limit:" + ip

		count, err := rateLimitScript.Run(
			ctx,
			redisClient,
			[]string{key},
			int(window.Seconds()),
		).Int()

		if err != nil {
			c.Next()
			return
		}

		if count > limit {
			c.JSON(
				http.StatusTooManyRequests,
				gin.H{"error": "too many requests"},
			)
			c.Abort()
			return
		}

		c.Next()
	}
}

// PlanRateLimiter enforces the per-user rate limit from their plan.
// Must run AFTER AuthMiddleware (requires org_id in context).
// Plan limits: FREE=100, PRO=500, Business=2000 req/min.
func PlanRateLimiter(redisClient *redis.Client) gin.HandlerFunc {

	const window = time.Minute

	return func(c *gin.Context) {

		// Read the plan rate limit set by auth middleware
		limitVal, exists := c.Get("rate_limit")
		if !exists {
			c.Next()
			return
		}

		limit, ok := limitVal.(int)
		if !ok || limit <= 0 {
			c.Next()
			return
		}

		orgIDVal, _ := c.Get("org_id")
		orgID, ok := orgIDVal.(uint)
		if !ok {
			c.Next()
			return
		}

		ctx := context.Background()
		key := "rate_plan:" + formatUint(orgID)

		count, err := rateLimitScript.Run(
			ctx,
			redisClient,
			[]string{key},
			int(window.Seconds()),
		).Int()

		if err != nil {
			c.Next()
			return
		}

		if count > limit {
			c.JSON(
				http.StatusTooManyRequests,
				gin.H{"error": "rate limit exceeded for your plan"},
			)
			c.Abort()
			return
		}

		c.Next()
	}
}

func formatUint(n uint) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
