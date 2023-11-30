package connector

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/samber/lo"
)

func ListMetrics(ctx context.Context, client client.BackendAPIClient, query models.GetMetricsRequest) (*models.GetMetricsResponse, error) {
	if len(query.Dimensions) == 0 {
		return nil, nil
	}
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
	return &models.GetMetricsResponse{
		Metrics: lo.Map(resp.GetMetrics(), func(m *pb.ListMetricsResponse_Metric, _ int) models.MetricDefinition {
			return models.MetricDefinition{
				Value:       m.Name,
				Label:       m.Name,
				Description: m.Description,
			}
		}),
	}, nil
}
