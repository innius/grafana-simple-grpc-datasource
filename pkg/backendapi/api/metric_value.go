package api

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/client"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
)

func valueQueryToInput(query models.MetricValueQuery) *pb.GetMetricValueRequest {
	var dimensions []*pb.Dimension
	for _, d := range query.Dimensions {
		dimensions = append(dimensions, &pb.Dimension{
			Key:   d.Key,
			Value: d.Value,
		})
	}

	metrics := make([]*pb.Metric, len(query.Metrics))
	for i := range query.Metrics {
		metrics[i] = &pb.Metric{
			Id: query.Metrics[i],
		}
	}
	return &pb.GetMetricValueRequest{
		Dimensions: dimensions,
		Metrics:    metrics,
	}
}

func GetMetricValue(ctx context.Context, client client.BackendAPIClient, query models.MetricValueQuery) (*framer.MetricValue, error) {
	clientReq := valueQueryToInput(query)

	resp, err := client.GetMetricValue(ctx, clientReq)

	if err != nil {
		return nil, err
	}

	return &framer.MetricValue{
		GetMetricValueResponse: resp,
	}, nil
}
