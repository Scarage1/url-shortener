package middleware

import (
	"net/http"
	"strings"

	"github.com/Scarage1/url-shortener/internal/model"
	"github.com/Scarage1/url-shortener/internal/service"
	"github.com/Scarage1/url-shortener/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware validates the JWT, extracts userID, resolves the user's
// primary organization, and sets user_id, org_id, email_verified, and
// rate_limit in the gin context.
func AuthMiddleware(jwtSecret string, orgService *service.OrgService, db *gorm.DB) gin.HandlerFunc {

	return func(c *gin.Context) {

		header :=
			c.GetHeader(
				"Authorization",
			)

		if header == "" {

			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "missing token",
				},
			)

			c.Abort()

			return
		}

		tokenString :=
			strings.TrimPrefix(
				header,
				"Bearer ",
			)

		userID, err :=
			utils.ValidateToken(
				tokenString,
				jwtSecret,
			)

		if err != nil {

			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid token",
				},
			)

			c.Abort()

			return
		}

		c.Set("user_id", userID)

		// Load user for email_verified flag
		var user model.User
		if err := db.First(&user, userID).Error; err == nil {
			c.Set("email_verified", user.EmailVerified)
		}

		// Resolve user → organization membership
		orgID, err := orgService.GetOrgIDForUser(userID)

		if err != nil {
			c.JSON(
				http.StatusForbidden,
				gin.H{
					"error": "no workspace found",
				},
			)
			c.Abort()
			return
		}

		c.Set("org_id", orgID)

		// Load org's plan rate limit for PlanRateLimiter
		var sub model.Subscription
		if err := db.Where(
			"organization_id = ? AND status = ?",
			orgID,
			model.SubscriptionActive,
		).Preload("Plan").First(&sub).Error; err == nil {
			c.Set("rate_limit", sub.Plan.RateLimit)
		}

		c.Next()
	}
}
