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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	pb "github.com/Albert-Ti/go-musthave-shortener-tpl/pkg/proto"
)

var currentToken string

func main() {
	ctx := context.Background()
	transportCreds, err := credentials.NewClientTLSFromFile("cert.pem", "")
	if err != nil {
		panic(err)
	}
	// Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(
		"127.0.0.1:3200",
		grpc.WithTransportCredentials(transportCreds),
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
	if err == nil {
		fmt.Println("ShortenURL result", resp.GetResult())
	} else {
		fmt.Printf("ShortenURL error: %v", err)
	}

	md := metadata.Pairs("authorization", currentToken)
	grpcCtx := metadata.NewOutgoingContext(context.Background(), md)

	resp2, err := c.ListUserURLs(grpcCtx, &pb.UserURLsRequest{})
	if err == nil {
		fmt.Println("ListUserURLs result", resp2.GetUrl())
	} else {
		fmt.Printf("ListUserURLs error: %v", err)
	}

	arr := strings.Split(resp.GetResult(), "/")
	id := arr[len(arr)-1]
	resp3, err := c.ExpandURL(grpcCtx, pb.URLExpandRequest_builder{Id: id}.Build())
	if err == nil {
		fmt.Println("ExpandURL result", resp3.GetResult())
	} else {
		fmt.Printf("ExpandURL error: %v", err)
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
