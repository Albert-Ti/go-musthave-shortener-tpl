package grpchandler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/audit"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/middleware"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"

	pb "github.com/Albert-Ti/go-musthave-shortener-tpl/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcServer struct {
	pb.UnimplementedShortenerServiceServer
	Server  *grpc.Server
	Auditor *audit.Auditor
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
	return nil, nil
}

func (g *GrpcServer) ListUserURLs(ctx context.Context, empty *emptypb.Empty) (*pb.UserURLsResponse, error) {
	return nil, nil
}

func (g *GrpcServer) Shutdown() {
	g.Server.GracefulStop()

	slog.Info("gRPC server shutdown gracefully")
}
