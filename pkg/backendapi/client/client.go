package client

import (
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto"
	"context"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
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
	res, err := b.backendAPI.ListDimensionKeys(ctx, in, opts...)
	if err != nil {
		return nil, handleError(err)
	}
	return res, nil
}

func (b *backendAPIClient) ListDimensionValues(ctx context.Context, in *pb.ListDimensionValuesRequest, opts ...grpc.CallOption) (*pb.ListDimensionValuesResponse, error) {
	res, err := b.backendAPI.ListDimensionValues(ctx, in, opts...)
	if err != nil {
		return nil, handleError(err)
	}
	return res, nil
}

func (b *backendAPIClient) ListMetrics(ctx context.Context, in *pb.ListMetricsRequest, opts ...grpc.CallOption) (*pb.ListMetricsResponse, error) {
	res, err := b.backendAPI.ListMetrics(ctx, in, opts...)
	if err != nil {
		return nil, handleError(err)
	}
	return res, nil
}

func (b *backendAPIClient) GetMetricValue(ctx context.Context, in *pb.GetMetricValueRequest, opts ...grpc.CallOption) (*pb.GetMetricValueResponse, error) {
	res, err := b.backendAPI.GetMetricValue(ctx, in, opts...)
	if err != nil {
		return nil, handleError(err)
	}
	return res, nil
}

func (b *backendAPIClient) GetMetricHistory(ctx context.Context, in *pb.GetMetricHistoryRequest, opts ...grpc.CallOption) (*pb.GetMetricHistoryResponse, error) {
	res, err := b.backendAPI.GetMetricHistory(ctx, in, opts...)
	if err != nil {
		return nil, handleError(err)
	}
	return res, nil
}

func (b *backendAPIClient) Dispose() {
	if err := b.conn.Close(); err != nil {
		log.DefaultLogger.Error("could not close connection on dispose", "error", err.Error())
	}
}


func handleError(err error) error {
	if err == nil {
		return nil
	}

	switch status.Convert(err).Code() {
	case codes.Canceled: {
		log.DefaultLogger.Error("server returned error with CANCELED status", "error", err.Error())
		return fmt.Errorf("request to server canceled")
	}
	case codes.DeadlineExceeded: {
		log.DefaultLogger.Error("server returned error with DEADLINE_EXCEEDED status", "error", err.Error())
		return fmt.Errorf("timeout")
	}
	case codes.NotFound: {
		log.DefaultLogger.Error("server returned error with NOT_FOUND status", "error", err.Error())
		return fmt.Errorf("server returned an error with NOT_FOUND status")
	}
	case codes.PermissionDenied: {
		log.DefaultLogger.Error("server returned error with PERMISSION_DENIED status", "error", err.Error())
		return fmt.Errorf("permission denied by server")
	}
	case codes.Unauthenticated: {
		log.DefaultLogger.Error("server returned an error with UNAUTHENTICATED status", "error", err.Error())
		return fmt.Errorf("server request has invalid authentication credentials")
	}

	default:
		log.DefaultLogger.Error("server returned an error", "code", status.Convert(err).Code(), "error", err)
		return fmt.Errorf("internal error",)
	}
}
