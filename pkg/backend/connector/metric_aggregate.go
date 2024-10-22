package connector

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"google.golang.org/protobuf/types/known/timestamppb"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	"github.com/samber/lo"
)

func aggregateQueryToInput(query models.MetricAggregateQuery) (*pb.GetMetricAggregateRequest, error) {
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
	return &pb.GetMetricAggregateRequest{
		IntervalMs:    query.Interval.Milliseconds(),
		MaxItems:      query.MaxDataPoints,
		Dimensions:    dimensions,
		Metrics:       metrics,
		StartDate:     timestamppb.New(query.TimeRange.From),
		EndDate:       timestamppb.New(query.TimeRange.To),
		StartingToken: query.NextToken,
		Options:       lo.MapValues(query.Options, func(value models.OptionValue, key string) string { return value.Value }),
	}, nil
}

func GetMetricAggregate(ctx context.Context, client client.BackendAPIClient, query models.MetricAggregateQuery) (*framer.MetricAggregate, error) {
	clientReq, err := aggregateQueryToInput(query)
	if err != nil {
		return nil, err
	}

	frames := map[string]*pb.Frame{}
	for {
		resp, err := client.GetMetricAggregate(ctx, clientReq)

		if err != nil {
			return nil, err
		}

		appendMatchingFrames(frames, resp.Frames)

		if resp != nil && resp.NextToken != "" {
			clientReq.StartingToken = resp.NextToken
			continue
		}
		break
	}
	return &framer.MetricAggregate{
		GetMetricAggregateResponse: &pb.GetMetricAggregateResponse{
			Frames: lo.MapToSlice(frames, func(_ string, v *pb.Frame) *pb.Frame { return v }),
		},
		Query: query.MetricBaseQuery,
	}, nil
}
