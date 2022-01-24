package factory

import (
	v1client "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/client/v1"
	v2client "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/client/v2"
	v2 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func NewClient(conn *grpc.ClientConn) (v2.GrafanaQueryAPIClient, error) {
	stub := rpb.NewServerReflectionClient(conn)

	c := grpcreflect.NewClient(context.Background(), stub)
	_, err := c.ResolveService("grafanav2.GrafanaQueryAPI")
	if err == nil {
		backend.Logger.Info("use v2 version of the backend API")
		return v2client.NewClient(conn)
	}
	backend.Logger.Info("use default version of the backend API")
	return v1client.NewClient(conn)
}
