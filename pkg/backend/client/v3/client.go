package v3

import (
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	v4 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v4"
	"google.golang.org/grpc"
)

func NewClient(conn *grpc.ClientConn) (v4.GrafanaQueryAPIClient, error) {
	return &adapter{adaptee: v3.NewGrafanaQueryAPIClient(conn)}, nil
}
