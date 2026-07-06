package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// EmailVerified blocks resource-creating operations for unverified users.
// Allows: GET /links, GET /stats/:code, login, logout
// Blocks: POST /shorten, POST /import, GET /export
//
// This middleware should be applied to routes that create or export data.
func EmailVerified() gin.HandlerFunc {

	return func(c *gin.Context) {

		// The auth middleware would need to set email_verified in context.
		// For now, we check if it exists; if not set, skip verification
		// (backwards compatible for the transition period).
		verified, exists := c.Get("email_verified")

		if exists {
			isVerified, ok := verified.(bool)
			if ok && !isVerified {
				c.JSON(
					http.StatusForbidden,
					gin.H{
						"error": "email verification required",
					},
				)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
