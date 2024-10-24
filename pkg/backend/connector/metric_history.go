package connector

import (
	"context"

	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
)

func historyQueryToInput(query models.MetricHistoryQuery) *pb.GetMetricHistoryRequest {
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
	return &pb.GetMetricHistoryRequest{
		Dimensions:    dimensions,
		Metrics:       metrics,
		StartDate:     timestamppb.New(query.TimeRange.From),
		EndDate:       timestamppb.New(query.TimeRange.To),
		StartingToken: query.NextToken,
		Options:       lo.MapValues(query.Options, func(value models.OptionValue, key string) string { return value.Value }),
	}
}

func GetMetricHistory(ctx context.Context, client client.BackendAPIClient, query models.MetricHistoryQuery) (*framer.MetricHistory, error) {
	clientReq := historyQueryToInput(query)

	frames := map[string]*pb.Frame{}
	for {
		resp, err := client.GetMetricHistory(ctx, clientReq)

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

	return &framer.MetricHistory{
		GetMetricHistoryResponse: &pb.GetMetricHistoryResponse{
			Frames: lo.MapToSlice(frames, func(_ string, v *pb.Frame) *pb.Frame { return v }),
		},
		Query: query,
	}, nil
}
