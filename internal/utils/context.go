package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// GetUserID extracts and type-asserts the user_id set by AuthMiddleware.
// Returns an error instead of panicking if the key is missing or the wrong type.
func GetUserID(c *gin.Context) (uint, error) {

	userIDValue, exists := c.Get("user_id")

	if !exists {

		return 0, errors.New("user_id not found in context")
	}

	userID, ok := userIDValue.(uint)

	if !ok {

		return 0, errors.New("user_id has unexpected type")
	}

	return userID, nil
}
