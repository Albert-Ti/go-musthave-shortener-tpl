package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MyCustomClaims расширяет стандартные claims jwt.RegisteredClaims полем UserID,
// чтобы связать выданный токен с конкретным пользователем сервиса.
// generate:reset
type MyCustomClaims struct {
	jwt.RegisteredClaims
	UserID string
}

// createToken подписывает новый JWT алгоритмом HS256 с claim UserID и сроком.
func CreateToken(userID string, secretKey string) (string, error) {
	t := jwt.New(jwt.SigningMethodHS256)

	t.Claims = &MyCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		userID,
	}

	return t.SignedString([]byte(secretKey))
}
