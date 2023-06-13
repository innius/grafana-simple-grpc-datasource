package client

import (
	"time"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client/factory"
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type backendClient struct {
	conn *grpc.ClientConn
	v3.GrafanaQueryAPIClient
}

func (b *backendClient) Dispose() {
	if err := b.conn.Close(); err != nil {
		log.DefaultLogger.Error("could not close connection on dispose", "error", err.Error())
	}
}
func getTransportCredentials(s BackendAPIDatasourceSettings) []grpc.DialOption {
	if s.APIKey == "" {
		return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}
	log.DefaultLogger.Info("dial with api-key authentication", "endpoint", s.Endpoint)
	return []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(nil)),
		grpc.WithPerRPCCredentials(ApiKeyAuthenticator{ApiKey: s.APIKey}),
	}
}

func New(settings BackendAPIDatasourceSettings) (BackendAPIClient, error) {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithMax(settings.MaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(500*time.Millisecond, 0.10)),
		grpc_retry.WithCodes(codes.ResourceExhausted),
	}
	options := getTransportCredentials(settings)
	options = append(options,
		grpc.WithUnaryInterceptor(GRPCDebugLogger()),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...)))

	conn, err := grpc.Dial(settings.Endpoint, options...)
	if err != nil {
		log.DefaultLogger.Error("could not dial")
		return nil, err
	}

	c, err := factory.NewClient(conn)
	if err != nil {
		return nil, err
	}
	return &backendClient{
		conn:                  conn,
		GrafanaQueryAPIClient: c,
	}, nil
}
