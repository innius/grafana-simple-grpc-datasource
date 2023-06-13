package connector

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
)

func ListMetrics(ctx context.Context, client client.BackendAPIClient, query models.MetricsQuery) (*framer.Metrics, error) {
	dimensions := make([]*pb.Dimension, len(query.Dimensions))
	for i, d := range query.Dimensions {
		dimensions[i] = &pb.Dimension{
			Key:   d.Key,
			Value: d.Value,
		}
	}
	resp, err := client.ListMetrics(ctx, &pb.ListMetricsRequest{
		Dimensions: dimensions,
		Filter:     query.Filter,
	})

	if err != nil {
		return nil, err
	}
	return &framer.Metrics{
		ListMetricsResponse: pb.ListMetricsResponse{
			Metrics: resp.Metrics,
		},
	}, nil
}
