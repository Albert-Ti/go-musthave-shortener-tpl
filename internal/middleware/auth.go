package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	mytoken "github.com/Albert-Ti/go-musthave-shortener-tpl/internal/token"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

type UserIDType string

// UserIDKey — ключ контекста.
const UserIDKey UserIDType = "userID"

var countUsers map[string]struct{}

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
func AuthGuard(secretKey string, isNotPostgres bool) func(http.Handler) http.Handler {
	if isNotPostgres {
		countUsers = map[string]struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCookie, err := r.Cookie("token")
			var authorizedUserID string

			if err != nil {
				slog.Error("cookie error", "error", err)
				authorizedUserID = utils.GenerateUUID()

				tokenString, err := mytoken.CreateToken(authorizedUserID, secretKey)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				newCookie := createCookie("token", tokenString)
				http.SetCookie(w, newCookie)

			} else {
				claims := &mytoken.MyCustomClaims{}

				token, err := jwt.ParseWithClaims(
					reqCookie.Value,
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

			if isNotPostgres {
				if _, ok := countUsers[authorizedUserID]; !ok {
					countUsers[authorizedUserID] = struct{}{}
				}
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

func GetCountUsers() int {
	return len(countUsers)
}
