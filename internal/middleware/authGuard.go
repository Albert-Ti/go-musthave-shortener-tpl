package middleware

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MyCustomClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

var secretKey = []byte("your-secret-key")

func createToken(userID string) (string, error) {

	t := jwt.New(jwt.SigningMethodHS256)

	t.Claims = &MyCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		userID,
	}

	token, err := t.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

func AuthGuard(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		next.ServeHTTP(w, r)
	}
}
