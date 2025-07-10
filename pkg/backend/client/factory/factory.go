package factory

import (
	"context"

	v1client "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client/v1"
	v2client "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client/v2"
	v3client "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client/v3"
	v4client "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client/v4"
	v4 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v4"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func NewClient(conn *grpc.ClientConn) (v4.GrafanaQueryAPIClient, error) {
	stub := rpb.NewServerReflectionClient(conn)
	c := grpcreflect.NewClient(context.Background(), stub)

	type versionInfo struct {
		service     string
		logMsg      string
		newClientFn func(*grpc.ClientConn) (v4.GrafanaQueryAPIClient, error)
	}

	versions := []versionInfo{
		{"grafanav4.GrafanaQueryAPI", "use v4 version of the backend API", v4client.NewClient},
		{"grafanav3.GrafanaQueryAPI", "use v3 version of the backend API", v3client.NewClient},
		{"grafanav2.GrafanaQueryAPI", "use v2 version of the backend API", v2client.NewClient},
	}

	for _, v := range versions {
		if descr, err := c.ResolveService(v.service); err == nil {
			backend.Logger.Info(v.logMsg, "descr", descr)
			return v.newClientFn(conn)
		}
	}

	backend.Logger.Info("use default version of the backend API")
	return v1client.NewClient(conn)
}
