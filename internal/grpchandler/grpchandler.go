package grpchandler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"

	pb "github.com/Albert-Ti/go-musthave-shortener-tpl/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	pb.UnimplementedShortenerServiceServer
	Server  *grpc.Server
	Svc     *service.Service
	BaseURL string
}

func (g *GrpcServer) ShortenURL(ctx context.Context, in *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	var response pb.URLShortenResponse

	userID, err := middleware.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthenticated")
	}

	keyURL, isNew, err := g.Svc.Save(ctx, in.GetUrl(), userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server")
	}

	if !isNew {
		return nil, status.Errorf(codes.AlreadyExists, "already exists")
	}

	response.SetResult(fmt.Sprintf("%s/%s", g.BaseURL, keyURL))

	return &response, nil
}

func (g *GrpcServer) ExpandURL(ctx context.Context, in *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	url, err := g.Svc.Get(ctx, in.GetId())

	if err != nil {
		if errors.Is(err, repository.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		if errors.Is(err, repository.ErrStatusGone) {
			return nil, status.Error(codes.FailedPrecondition, "status gone")
		}
		return nil, status.Error(codes.Internal, "internal server")
	}

	response := *pb.URLExpandResponse_builder{
		Result: url,
	}.Build()

	return &response, nil
}

func (g *GrpcServer) ListUserURLs(ctx context.Context, in *pb.UserURLsRequest) (*pb.UserURLsResponse, error) {
	userID, err := middleware.GetAuthUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthenticated")
	}
	urls, err := g.Svc.GetAll(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal server")
	}

	if len(urls) == 0 {
		return nil, status.Errorf(codes.NotFound, "not found")
	}
	var list []*pb.URLData

	for _, v := range urls {
		list = append(list, pb.URLData_builder{
			ShortUrl:    v.ShortURL,
			OriginalUrl: v.OriginalURL,
		}.Build())
	}

	response := pb.UserURLsResponse_builder{
		Url: list,
	}.Build()

	return response, nil
}

func (g *GrpcServer) Shutdown() {
	g.Server.GracefulStop()

	slog.Info("gRPC server shutdown gracefully")
}
