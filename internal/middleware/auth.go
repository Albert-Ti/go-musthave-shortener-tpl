package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

// UserIDKey — ключ контекста.
type UserIDType string

const UserIDKey UserIDType = "userID"

// MyCustomClaims расширяет стандартные claims jwt.RegisteredClaims полем UserID,
// чтобы связать выданный токен с конкретным пользователем сервиса.
type MyCustomClaims struct {
	jwt.RegisteredClaims
	UserID string
}

// createToken подписывает новый JWT алгоритмом HS256 с claim UserID и сроком.
func createToken(userID string, secretKey string) (string, error) {
	t := jwt.New(jwt.SigningMethodHS256)

	t.Claims = &MyCustomClaims{
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		userID,
	}

	return t.SignedString([]byte(secretKey))
}

// createCookie создаёт HttpOnly-cookie с заданными именем и значением.
func createCookie(name string, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
	}
}

// AuthGuard возвращает middleware, которое аутентифицирует пользователя по
// JWT-токену, хранящемуся в cookie "token".
//
// Если cookie отсутствует, middleware генерирует новый UserID (utils.GenerateUUID),
// выпускает для него JWT, подписанный secretKey, и устанавливает его в cookie
// "token" в ответе — так на первый запрос анонимный пользователь получает
// постоянный идентификатор.
//
// Если cookie присутствует, middleware проверяет подпись и валидность токена.
// При неверной подписи, истёкшем сроке действия или пустом UserID в claims
// запрос прерывается с 401 Unauthorized.
//
// Пример использования:
//
//	r := chi.NewRouter()
//	r.Use(middleware.AuthGuard(cfg.JWTSecretKey))
func AuthGuard(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			var authorizedUserID string

			if err != nil {
				slog.Error("cookie error", "error", err)
				authorizedUserID = utils.GenerateUUID()

				tokenString, err := createToken(authorizedUserID, secretKey)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				cookie := createCookie("token", tokenString)
				http.SetCookie(w, cookie)
			} else {
				claims := &MyCustomClaims{}

				token, err := jwt.ParseWithClaims(
					cookie.Value,
					claims,
					func(t *jwt.Token) (any, error) {
						return []byte(secretKey), nil
					},
				)

				if err != nil || !token.Valid || claims.UserID == "" {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				authorizedUserID = claims.UserID
			}

			ctx := context.WithValue(r.Context(), UserIDKey, authorizedUserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAuthUserID извлекает идентификатор пользователя.
func GetAuthUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("user id not found")
	}
	return userID, nil
}
