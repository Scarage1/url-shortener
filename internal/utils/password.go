package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// BcryptCost is the bcrypt work factor.
// Production default is 14. Tests override it to 10 via TestMain for speed.
var BcryptCost = 14

func HashPassword(
	password string,
) (string, error) {

	bytes, err :=
		bcrypt.GenerateFromPassword(
			[]byte(password),
			BcryptCost,
		)

	return string(bytes), err
}

func CheckPassword(
	password string,
	hash string,
) bool {

	err :=
		bcrypt.CompareHashAndPassword(
			[]byte(hash),
			[]byte(password),
		)

	return err == nil
}
