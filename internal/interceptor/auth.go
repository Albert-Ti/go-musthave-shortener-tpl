package interceptor

import (
	"context"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	mytoken "github.com/Albert-Ti/go-musthave-shortener-tpl/internal/token"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"github.com/golang-jwt/jwt/v5"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthGuard(secretKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var tokenStr string
		var authorizedUserID string

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get("authorization")
			if len(values) > 0 {
				tokenStr = values[0]

				claims := &mytoken.MyCustomClaims{}

				token, err := jwt.ParseWithClaims(
					tokenStr,
					claims,
					func(t *jwt.Token) (any, error) {
						return []byte(secretKey), nil
					},
				)

				if err != nil || !token.Valid || claims.UserID == "" {
					return nil, status.Error(codes.Unauthenticated, "token no valid")
				}

				authorizedUserID = claims.UserID
			}
		}

		if tokenStr == "" {
			authorizedUserID = utils.GenerateUUID()

			newToken, err := mytoken.CreateToken(authorizedUserID, secretKey)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to create token")
			}

			_ = grpc.SetHeader(ctx, metadata.Pairs("authorization", newToken))
		}
		ctx = middleware.SetAuthUserID(ctx, authorizedUserID)
		return handler(ctx, req)
	}
}
