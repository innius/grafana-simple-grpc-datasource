package api

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	"context"
	"fmt"
	"strings"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
)

func aggregateQueryToInput(query models.MetricAggregateQuery) (*pb.GetMetricAggregateRequest, error) {
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

	aggType, err := parseAggregateType(query.AggregateType)
	if err != nil {
		return nil, err
	}
	return &pb.GetMetricAggregateRequest{
		Dimensions:    dimensions,
		Metrics:       metrics,
		StartDate:     query.TimeRange.From.Unix(),
		EndDate:       query.TimeRange.To.Unix(),
		AggregateType: aggType,
		StartingToken: query.NextToken,
	}, nil

}

func parseAggregateType(s string) (pb.AggregateType, error) {
	switch strings.ToLower(s) {
	case strings.ToLower(pb.AggregateType_AVERAGE.String()):
		return pb.AggregateType_AVERAGE, nil
	case strings.ToLower(pb.AggregateType_MAX.String()):
		return pb.AggregateType_MAX, nil
	case strings.ToLower(pb.AggregateType_MIN.String()):
		return pb.AggregateType_MIN, nil
	default:
		var t pb.AggregateType
		return t, fmt.Errorf("aggregate type %s is not supported by backend plugin", s)
	}
}

func GetMetricAggregate(ctx context.Context, client client.BackendAPIClient, query models.MetricAggregateQuery) (*framer.MetricAggregate, error) {

	clientReq, err := aggregateQueryToInput(query)
	if err != nil {
		return nil, err
	}

	resp, err := client.GetMetricAggregate(ctx, clientReq)

	if err != nil {
		return nil, err
	}
	return &framer.MetricAggregate{
		GetMetricAggregateResponse: &pb.GetMetricAggregateResponse{
			Result:    resp.Result,
			NextToken: resp.NextToken,
		},
		AggregationType: clientReq.AggregateType,
	}, nil
}
