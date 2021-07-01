package api

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto"
)

func historyQueryToInput(query models.MetricHistoryQuery) *pb.GetMetricHistoryRequest {
	var dimensions []*pb.Dimension
	for _, d := range query.Dimensions {
		dimensions = append(dimensions, &pb.Dimension{
			Key:   d.Key,
			Value: d.Value,
		})
	}
	return &pb.GetMetricHistoryRequest{
		Dimensions:    dimensions,
		Metric:        query.MetricId,
		StartDate:     query.TimeRange.From.Unix(),
		EndDate:       query.TimeRange.To.Unix(),
		StartingToken: query.NextToken,
	}
}

func GetMetricHistory(ctx context.Context, client client.BackendAPIClient, query models.MetricHistoryQuery) (*framer.MetricHistory, error) {
	clientReq := historyQueryToInput(query)

	resp, err := client.GetMetricHistory(ctx, clientReq)

	if err != nil {
		return nil, err
	}
	return &framer.MetricHistory{
		GetMetricHistoryResponse: pb.GetMetricHistoryResponse{
			Values:    resp.Values,
			NextToken: resp.NextToken,
		},
		MetricID: query.MetricId,
	}, nil
}
