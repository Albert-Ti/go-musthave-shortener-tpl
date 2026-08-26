package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Albert-Ti/go-musthave-shortener-tpl/pkg/proto"
)

func main() {
	ctx := context.Background()

	// Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(
		`:3200`,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(clientInterceptor),
	)

	if err != nil {
		slog.Error("ошибка при установлении соединения с сервером", "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	c := pb.NewShortenerServiceClient(conn)

	resp, err := c.ShortenURL(ctx, pb.URLShortenRequest_builder{
		Url: "https://example.com",
	}.Build())

	fmt.Println(resp.GetResult())

	if err != nil {
		fmt.Printf("ошибка при получении информации о пользователе: %w", err)
	}
}

func clientInterceptor(
	ctx context.Context, method string, req any, reply any,
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	// выполняем действия перед вызовом метода
	start := time.Now()

	// вызываем RPC-метод
	err := invoker(ctx, method, req, reply, cc, opts...)

	// выполняем действия после вызова метода
	if err != nil {
		log.Printf("[ERROR] %s,%v", method, err)
	} else {
		log.Printf("[INFO] %s,%v", method, time.Since(start))
	}
	return err
}
