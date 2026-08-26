package interceptor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor - интерцептор для логирования gRPC запросов
func Logging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Выполняем запрос
		resp, err := handler(ctx, req)

		// Получаем статус
		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			} else {
				statusCode = codes.Unknown
			}
		}

		// Длительность запроса
		duration := time.Since(start)

		// Логируем в зависимости от статуса
		method := info.FullMethod
		logger := slog.With(
			slog.String("method", method),
			slog.Duration("duration", duration),
			slog.String("code", statusCode.String()),
			slog.Int("code_value", int(statusCode)),
		)

		fmt.Println(statusCode)
		switch {
		case statusCode >= codes.Internal: // 5xx
			logger.Error("gRPC request failed",
				slog.String("error", err.Error()),
			)
		case statusCode >= codes.InvalidArgument: // 4xx
			logger.Warn("gRPC request failed",
				slog.String("error", err.Error()),
			)
		default:
			logger.Info("gRPC")
		}

		return resp, err
	}
}
