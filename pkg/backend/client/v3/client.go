package v2

import (
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"google.golang.org/grpc"
)

func NewClient(conn *grpc.ClientConn) (v3.GrafanaQueryAPIClient, error) {
	return v3.NewGrafanaQueryAPIClient(conn), nil
}
