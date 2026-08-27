package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Albert-Ti/go-musthave-shortener-tpl/pkg/proto"
)

var currentToken string

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

	fmt.Println("ShortenURL result", resp.GetResult())

	md := metadata.Pairs("authorization", currentToken)
	grpcCtx := metadata.NewOutgoingContext(context.Background(), md)

	resp2, err := c.ListUserURLs(grpcCtx, &emptypb.Empty{})

	fmt.Println("ListUserURLs result", resp2.GetUrl())

	arr := strings.Split(resp.GetResult(), "/")
	id := arr[len(arr)-1]

	resp3, err := c.ExpandURL(grpcCtx, pb.URLExpandRequest_builder{
		Id: id,
	}.Build())

	fmt.Println("ExpandURL result", resp3.GetResult())

	if err != nil {
		fmt.Printf("ошибка при получении списка ссылок пользователя: %v", err)
	}
}

func clientInterceptor(
	ctx context.Context, method string, req any, reply any,
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	var header metadata.MD
	opts = append(opts, grpc.Header(&header)) // ловим заголовки ответа

	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)

	if values := header.Get("authorization"); len(values) > 0 {
		fmt.Println("получен новый токен:", values[0])
		currentToken = values[0]
	}

	if err != nil {
		log.Printf("[ERROR] %s,%v", method, err)
	} else {
		log.Printf("[INFO] %s,%v", method, time.Since(start))
	}
	return err
}
