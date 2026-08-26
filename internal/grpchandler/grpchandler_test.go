package grpchandler_test

import (
	"context"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/Albert-Ti/go-musthave-shortener-tpl/pkg/proto"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/audit"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/grpchandler"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/interceptor"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository/mocks"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
)

const bufSize = 1024 * 1024

func newTestGRPCServer(t *testing.T, svc *service.Service, auditor *audit.Auditor, opts *config.Options) pb.ShortenerServiceClient {
	t.Helper()

	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.Auth(opts.JWTSecretKey),
			interceptor.Logging(),
		),
	)
	gs := &grpchandler.GrpcServer{Server: srv, Svc: svc, Auditor: auditor}
	pb.RegisterShortenerServiceServer(srv, gs)

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Logf("bufconn server exited: %v", err)
		}
	}()
	t.Cleanup(srv.Stop)

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return pb.NewShortenerServiceClient(conn)
}

func TestCreateURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)

	mockRepo.EXPECT().
		Save(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1)

	opts := config.NewOptions(config.WithBaseURL("http://localhost:8080"))
	svc := service.NewService(mockRepo, opts)
	auditor, err := audit.NewAuditor("", "", 1, 1)

	require.NoError(t, err)

	client := newTestGRPCServer(t, svc, auditor, opts)

	req := pb.URLShortenRequest_builder{Url: "https://example.com"}.Build()
	resp, err := client.ShortenURL(context.Background(), req)

	require.NoError(t, err)
	require.NotEmpty(t, resp.GetResult())
}
