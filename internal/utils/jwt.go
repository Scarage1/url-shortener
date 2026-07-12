package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken generates a short-lived access token valid for 15 minutes.
func GenerateToken(
	userID uint,
	secret string,
) (string, error) {

	claims :=
		jwt.MapClaims{
			"user_id": userID,
			"exp": time.Now().
				Add(
					time.Minute * 15,
				).
				Unix(),
		}

	token :=
		jwt.NewWithClaims(
			jwt.SigningMethodHS256,
			claims,
		)

	return token.SignedString(
		[]byte(secret),
	)
}

func ValidateToken(
	tokenString string,
	secret string,
) (uint, error) {

	token, err :=
		jwt.Parse(
			tokenString,
			func(token *jwt.Token) (interface{}, error) {

				if _, ok :=
					token.Method.(*jwt.SigningMethodHMAC); !ok {

					return nil,
						jwt.ErrTokenSignatureInvalid
				}

				return []byte(secret), nil
			},
		)

	if err != nil {
		return 0, err
	}

	if claims, ok :=
		token.Claims.(jwt.MapClaims); ok &&
		token.Valid {

		userIDFloat, ok :=
			claims["user_id"].(float64)

		if !ok {
			return 0,
				jwt.ErrTokenInvalidClaims
		}

		return uint(userIDFloat), nil
	}

	return 0, jwt.ErrTokenInvalidClaims
}

// GenerateRefreshToken generates a secure, random 32-byte hex-encoded string.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashRefreshToken SHA-256 hashes a high-entropy refresh token.
func HashRefreshToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}
