package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDKey string = "userID"

var secretKey = []byte(config.Envs.JWTSecretKey)

type MyCustomClaims struct {
	jwt.RegisteredClaims
	UserID string
}

func createToken(userID string) (string, error) {
	t := jwt.New(jwt.SigningMethodHS256)

	t.Claims = &MyCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		userID,
	}

	return t.SignedString(secretKey)
}

func createCookie(name string, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
	}
}

func AuthGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		var authorizedUserID string

		if err != nil {
			slog.Error("cookie error", "error", err)
			authorizedUserID = utils.GenerateUUID()

			token, err := createToken(authorizedUserID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			cookie := createCookie("token", token)
			http.SetCookie(w, cookie)
		} else {
			claims := &MyCustomClaims{}
			_, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (any, error) {
				return secretKey, nil
			})
			if err != nil {
				slog.Error("jwt parse error", "error", err)
			}
			if claims.UserID == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			authorizedUserID = claims.UserID
		}

		ctx := context.WithValue(r.Context(), UserIDKey, authorizedUserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
