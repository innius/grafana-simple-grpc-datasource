package v4

import (
	v4 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v4"
	"google.golang.org/grpc"
)

func NewClient(conn *grpc.ClientConn) (v4.GrafanaQueryAPIClient, error) {
	return v4.NewGrafanaQueryAPIClient(conn), nil
}
