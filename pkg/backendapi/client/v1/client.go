package v1

import (
	v1 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v1"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"google.golang.org/grpc"
)

func NewClient(conn *grpc.ClientConn) (v2.GrafanaQueryAPIClient, error) {
	return &adapter{v1Client: v1.NewGrafanaQueryAPIClient(conn)}, nil
}
