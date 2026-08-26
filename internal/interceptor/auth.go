package interceptor

import (
	"context"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/token"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func Auth(secretKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		userID := utils.GenerateUUID()
		token, err := token.CreateToken(userID, secretKey)
		if err != nil {
			return nil, err
		}

		md := metadata.New(map[string]string{"token": token})
		ctx = metadata.NewOutgoingContext(ctx, md)

		return handler(ctx, req)
	}

}
