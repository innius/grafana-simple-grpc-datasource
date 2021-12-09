package v2

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"google.golang.org/grpc"
)

func NewClient(conn *grpc.ClientConn) (v2.GrafanaQueryAPIClient, error) {
	return v2.NewGrafanaQueryAPIClient(conn), nil
}
