package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)


func GenerateToken(
	userID uint,
	secret string,
) (string, error) {

	claims :=
		jwt.MapClaims{

			"user_id": userID,

			"exp": time.Now().
				Add(
					time.Hour * 24,
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

				return []byte(secret), nil
			},
		)

	if err != nil {

		return 0, err
	}

	if claims, ok :=
		token.Claims.(jwt.MapClaims); ok &&
		token.Valid {

		userID :=
			uint(
				claims["user_id"].(float64),
			)

		return userID, nil
	}

	return 0, jwt.ErrTokenInvalidClaims
}
