package client

import (
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto"
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type BackendAPIClient interface {
	pb.GrafanaQueryAPIClient
	Dispose()
}

type backendAPIClient struct {
	backendAPI pb.GrafanaQueryAPIClient
	conn       *grpc.ClientConn
}

func NewClient(settings BackendAPIDatasourceSettings) (BackendAPIClient, error) {
	options := []grpc.DialOption{}
	if settings.ApiKeyAuthenticationEnabled {
		log.DefaultLogger.Info("dial with api-key authentication", "endpoint", settings.Endpoint)
		options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(nil)),
			grpc.WithPerRPCCredentials(apikeyAuthentication{
				apiKey: settings.APIKey,
			}),
			grpc.WithUnaryInterceptor(GRPCDebugLogger()),
		)
	} else {
		log.DefaultLogger.Info("dial without credentials", "endpoint", settings.Endpoint)
		options = append(options, grpc.WithUnaryInterceptor(GRPCDebugLogger()),
			grpc.WithInsecure(),)
	}

	conn, err := grpc.Dial(settings.Endpoint, options...)
	if err != nil {
		log.DefaultLogger.Error("could not dial")
		return nil, err
	}

	return &backendAPIClient{conn: conn, backendAPI: pb.NewGrafanaQueryAPIClient(conn)}, nil
}

func (b *backendAPIClient) ListDimensionKeys(ctx context.Context, in *pb.ListDimensionKeysRequest, opts ...grpc.CallOption) (*pb.ListDimensionKeysResponse, error) {
	return b.backendAPI.ListDimensionKeys(ctx, in, opts...)
}

func (b *backendAPIClient) ListDimensionValues(ctx context.Context, in *pb.ListDimensionValuesRequest, opts ...grpc.CallOption) (*pb.ListDimensionValuesResponse, error) {
	return b.backendAPI.ListDimensionValues(ctx, in, opts...)
}

func (b *backendAPIClient) ListMetrics(ctx context.Context, in *pb.ListMetricsRequest, opts ...grpc.CallOption) (*pb.ListMetricsResponse, error) {
	return b.backendAPI.ListMetrics(ctx, in, opts...)
}

func (b *backendAPIClient) GetMetricValue(ctx context.Context, in *pb.GetMetricValueRequest, opts ...grpc.CallOption) (*pb.GetMetricValueResponse, error) {
	return b.backendAPI.GetMetricValue(ctx, in, opts...)
}

func (b *backendAPIClient) GetMetricHistory(ctx context.Context, in *pb.GetMetricHistoryRequest, opts ...grpc.CallOption) (*pb.GetMetricHistoryResponse, error) {
	return b.backendAPI.GetMetricHistory(ctx, in, opts...)
}

func (b *backendAPIClient) Dispose() {
	if err := b.conn.Close(); err != nil {
		log.DefaultLogger.Error("could not close connection on dispose", "error", err.Error())
	}
}
