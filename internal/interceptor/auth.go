package interceptor

import (
	"context"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/token"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Auth(secretKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		var tokenStr string
		if ok {
			values := md.Get("authorization")
			if len(values) > 0 {
				tokenStr = values[0]
			}
		}
		var userID string
		if tokenStr == "" {
			userID = utils.GenerateUUID()

			newToken, err := token.CreateToken(userID, secretKey)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to create token")
			}

			_ = grpc.SetHeader(ctx, metadata.Pairs("authorization", newToken))
		}
		ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		return handler(ctx, req)
	}

}
