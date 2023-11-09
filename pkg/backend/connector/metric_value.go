package connector

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		Options:    lo.MapValues(query.Options, func(value models.OptionValue, key string) string { return value.Value }),
		Dimensions: dimensions,
		Metrics:    metrics,
		StartDate:  timestamppb.New(query.TimeRange.From),
		EndDate:    timestamppb.New(query.TimeRange.To),
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
		Query:                  query,
	}, nil
}
