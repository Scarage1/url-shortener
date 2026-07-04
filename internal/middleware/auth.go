package middleware

import (
	"net/http"
	"strings"

	"github.com/Scarage1/url-shortener/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {

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

		c.Set(
			"user_id",
			userID,
		)

		c.Next()
	}
}
