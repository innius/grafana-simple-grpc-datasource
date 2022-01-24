package connector

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"context"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"

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
	metrics := make([]string, len(query.Metrics))
	for i := range query.Metrics {
		metrics[i] = query.Metrics[i].MetricId
	}
	return &pb.GetMetricValueRequest{
		Dimensions: dimensions,
		Metric:     metrics,
	}
}

func GetMetricValue(ctx context.Context, client client.BackendAPIClient, query models.MetricValueQuery) (*framer.MetricValue, error) {
	clientReq := valueQueryToInput(query)

	resp, err := client.GetMetricValue(ctx, clientReq)

	if err != nil {
		return nil, err
	}

	backend.Logger.Info(fmt.Sprintf("the response %+v", resp))
	return &framer.MetricValue{
		GetMetricValueResponse: pb.GetMetricValueResponse{
			Data: resp.Data,
		},
		MetricValueQuery: query,
	}, nil
}
