package v1

import (
	v1 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v1"
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"google.golang.org/grpc"
)

func NewClient(conn *grpc.ClientConn) (v3.GrafanaQueryAPIClient, error) {
	return &adapter{v1Client: v1.NewGrafanaQueryAPIClient(conn)}, nil
}
