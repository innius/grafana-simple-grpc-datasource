package api

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
)

func valueQueryToInput(query models.MetricValueQuery) *pb.GetMetricValueRequest {
	var dimensions []*pb.Dimension
	for _, d := range query.Dimensions {
		dimensions = append(dimensions, &pb.Dimension{
			Key:   d.Key,
			Value: d.Value,
		})
	}
	return &pb.GetMetricValueRequest{
		Dimensions: dimensions,
		Metric:     query.MetricId,
	}
}

func GetMetricValue(ctx context.Context, client client.BackendAPIClient, query models.MetricValueQuery) (*framer.MetricValue, error) {
	clientReq := valueQueryToInput(query)

	resp, err := client.GetMetricValue(ctx, clientReq)

	if err != nil {
		return nil, err
	}
	return &framer.MetricValue{
		GetMetricValueResponse: pb.GetMetricValueResponse{
			Timestamp: resp.Timestamp,
			Values:     resp.Values,
		},
		MetricValueQuery: query,
	}, nil
}
