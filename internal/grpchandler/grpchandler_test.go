package grpchandler_test

import (
	"context"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/Albert-Ti/go-musthave-shortener-tpl/pkg/proto"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/grpchandler"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/interceptor"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository/mocks"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/token"
)

const bufSize = 1024 * 1024

func newTestGRPCServer(t *testing.T, svc *service.Service, opts *config.Options) pb.ShortenerServiceClient {
	t.Helper()

	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.AuthGuard(opts.JWTSecretKey),
			interceptor.Logging(),
		),
	)
	gs := &grpchandler.GrpcServer{Server: srv, Svc: svc}
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

func TestShortenURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)

	opts := config.NewOptions(config.WithGRPCRunAddr("localhost:3200"))
	svc := service.NewService(mockRepo, opts)

	client := newTestGRPCServer(t, svc, opts)

	validToken, err := token.CreateToken("user-1", opts.JWTSecretKey)
	require.NoError(t, err)

	tests := []struct {
		name      string
		wantCode  codes.Code
		setupMock func(mock *mocks.MockRepository)
	}{
		{
			name:     "Case OK",
			wantCode: codes.OK,
			setupMock: func(mock *mocks.MockRepository) {
				mockRepo.EXPECT().
					Save(gomock.Any(), gomock.Any(), "https://example.com", "user-1").
					Return("key_1", nil).
					Times(1)
			},
		},

		{
			name:     "Case already exists",
			wantCode: codes.AlreadyExists,
			setupMock: func(mock *mocks.MockRepository) {
				mockRepo.EXPECT().
					Save(gomock.Any(), gomock.Any(), "https://example.com", "user-1").
					Return("", repository.ErrConflict).
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := pb.URLShortenRequest_builder{Url: "https://example.com"}.Build()

			tt.setupMock(mockRepo)

			md := metadata.Pairs("authorization", validToken)
			grpcCtx := metadata.NewOutgoingContext(context.Background(), md)

			resp, err := client.ShortenURL(grpcCtx, req)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())

			if tt.name == "Case OK" {
				require.NotEmpty(t, resp.GetResult())

			}
		})
	}
}

func TestListUserURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)

	opts := config.NewOptions(
		config.WithGRPCRunAddr("localhost:3200"),
	)

	svc := service.NewService(mockRepo, opts)

	client := newTestGRPCServer(t, svc, opts)

	validToken, err := token.CreateToken("user-1", opts.JWTSecretKey)
	require.NoError(t, err)

	tests := []struct {
		name      string
		wantCode  codes.Code
		setupMock func(mock *mocks.MockRepository)
	}{
		{
			name:     "Case OK",
			wantCode: codes.OK,
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					GetAll(gomock.Any(), "user-1").
					Return([]map[string]string{{
						"key": "http://localhost:8080/key_1",
						"url": "https://google.com",
					},
					}, nil).Times(1)
			},
		},
		{
			name:     "Case NotFound",
			wantCode: codes.NotFound,
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					GetAll(gomock.Any(), "user-1").
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockRepo)

			md := metadata.Pairs("authorization", validToken)
			grpcCtx := metadata.NewOutgoingContext(context.Background(), md)
			resp, err := client.ListUserURLs(grpcCtx, &pb.UserURLsRequest{})

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())

			if tt.wantCode == codes.OK {
				require.NotEmpty(t, resp.GetUrl())
			}
		})
	}
}

func TestExpandURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepository(ctrl)

	opts := config.NewOptions(
		config.WithGRPCRunAddr("localhost:3200"),
	)

	svc := service.NewService(mockRepo, opts)

	client := newTestGRPCServer(t, svc, opts)

	validToken, err := token.CreateToken("user-1", opts.JWTSecretKey)
	require.NoError(t, err)

	tests := []struct {
		name      string
		wantCode  codes.Code
		setupMock func(mock *mocks.MockRepository)
	}{
		{
			name:     "Case OK",
			wantCode: codes.OK,
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					Get(gomock.Any(), "key_1").
					Return("http://example.com", nil).
					Times(1)
			},
		},
		{
			name:     "Case Not Found",
			wantCode: codes.NotFound,
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					Get(gomock.Any(), "key_1").
					Return("", repository.ErrNoRows).
					Times(1)
			},
		},

		{
			name:     "Case Status Gone",
			wantCode: codes.FailedPrecondition,
			setupMock: func(mock *mocks.MockRepository) {
				mock.EXPECT().
					Get(gomock.Any(), "key_1").
					Return("", repository.ErrStatusGone).
					Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock(mockRepo)

			md := metadata.Pairs("authorization", validToken)
			grpcCtx := metadata.NewOutgoingContext(context.Background(), md)

			req := pb.URLExpandRequest_builder{Id: "key_1"}.Build()
			resp, err := client.ExpandURL(grpcCtx, req)

			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, tt.wantCode, st.Code())

			if tt.wantCode == codes.OK {
				require.NotEmpty(t, resp.GetResult())
			}
		})
	}
}
