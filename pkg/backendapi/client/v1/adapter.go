package v1

import (
	v1 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v1"
	v2 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type adapter struct {
	v1Client v1.GrafanaQueryAPIClient
}

func (b *adapter) ListDimensionKeys(ctx context.Context, in *v2.ListDimensionKeysRequest, opts ...grpc.CallOption) (*v2.ListDimensionKeysResponse, error) {
	inv1 := &v1.ListDimensionKeysRequest{
		Filter: in.Filter,
	}
	res, err := b.v1Client.ListDimensionKeys(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	r := make([]*v2.ListDimensionKeysResponse_Result, len(res.Results))
	for i := range res.Results {
		r[i] = &v2.ListDimensionKeysResponse_Result{
			Key:         res.Results[i].Key,
			Description: res.Results[i].Description,
		}
	}
	return &v2.ListDimensionKeysResponse{
		Results: r,
	}, nil
}

func (b *adapter) ListDimensionValues(ctx context.Context, in *v2.ListDimensionValuesRequest, opts ...grpc.CallOption) (*v2.ListDimensionValuesResponse, error) {
	inv1 := &v1.ListDimensionValuesRequest{
		DimensionKey: in.DimensionKey,
		Filter:       in.Filter,
	}
	res, err := b.v1Client.ListDimensionValues(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	r := make([]*v2.ListDimensionValuesResponse_Result, len(res.Results))
	for i := range res.Results {
		r[i] = &v2.ListDimensionValuesResponse_Result{
			Value:       res.Results[i].Value,
			Description: res.Results[i].Description,
		}
	}
	return &v2.ListDimensionValuesResponse{
		Results: r,
	}, nil
}

func (b *adapter) ListMetrics(ctx context.Context, in *v2.ListMetricsRequest, opts ...grpc.CallOption) (*v2.ListMetricsResponse, error) {
	inv1 := &v1.ListMetricsRequest{
		Dimensions: toV1Dimensions(in.Dimensions),
		Filter:     in.Filter,
	}
	res, err := b.v1Client.ListMetrics(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	r := make([]*v2.ListMetricsResponse_Metric, len(res.Metrics))
	for i := range res.Metrics {
		r[i] = &v2.ListMetricsResponse_Metric{
			Name:        res.Metrics[i].Name,
			Description: res.Metrics[i].Description,
		}
	}
	return &v2.ListMetricsResponse{
		Metrics: r,
	}, nil
}

func (b *adapter) GetMetricValue(ctx context.Context, in *v2.GetMetricValueRequest, opts ...grpc.CallOption) (*v2.GetMetricValueResponse, error) {
	if len(in.Metrics) == 0 {
		return &v2.GetMetricValueResponse{}, nil
	}
	metricId := in.Metrics[0]
	inv1 := &v1.GetMetricValueRequest{
		Dimensions: toV1Dimensions(in.Dimensions),
		Metric:     metricId,
	}
	res, err := b.v1Client.GetMetricValue(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}

	var value float64
	if res.Value != nil {
		value = res.Value.DoubleValue
	}
	return &v2.GetMetricValueResponse{
		Frames: []*v2.GetMetricValueResponse_Frame{
			{
				Metric:    metricId,
				Timestamp: timestamppb.New(getTime(res.Timestamp)),
				Fields: []*v2.SingleValueField{
					{
						Name:   "",
						Labels: nil,
						Config: nil,
						Value:  value,
					},
				},
			},
		},
	}, nil
}

func toV1Dimensions(dims []*v2.Dimension) []*v1.Dimension {
	d := make([]*v1.Dimension, len(dims))
	for i := range dims {
		v := dims[i]
		d[i] = &v1.Dimension{
			Key:   v.Key,
			Value: v.Value,
		}
	}
	return d
}

func (b *adapter) GetMetricHistory(ctx context.Context, in *v2.GetMetricHistoryRequest, opts ...grpc.CallOption) (*v2.GetMetricHistoryResponse, error) {
	if len(in.Metrics) == 0 {
		return &v2.GetMetricHistoryResponse{}, nil
	}
	metricId := in.Metrics[0]
	inv1 := &v1.GetMetricHistoryRequest{
		Dimensions:    toV1Dimensions(in.Dimensions),
		Metric:        metricId,
		StartDate:     in.StartDate.AsTime().Unix(),
		EndDate:       in.EndDate.AsTime().Unix(),
		MaxItems:      in.MaxItems,
		TimeOrdering:  v1.TimeOrdering(in.TimeOrdering),
		StartingToken: in.StartingToken,
	}
	res, err := b.v1Client.GetMetricHistory(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	timestamps := make([]*timestamppb.Timestamp, len(res.Values))
	doubleValues := make([]float64, len(res.Values))
	for i := range res.Values {
		v := res.Values[i]
		if v == nil {
			continue
		}
		timestamps[i] = timestamppb.New(getTime(v.Timestamp))
		var value float64
		if v.Value != nil {
			value = v.Value.DoubleValue
		}
		doubleValues[i] = value
	}
	return &v2.GetMetricHistoryResponse{
		Frames: []*v2.Frame{
			{
				Metric:     metricId,
				Timestamps: timestamps,
				Fields: []*v2.Field{
					{
						Name:   "",
						Labels: nil,
						Config: nil,
						Values: doubleValues,
					},
				},
			},
		},
		NextToken: res.NextToken,
	}, nil
}

func (b *adapter) GetMetricAggregate(ctx context.Context, in *v2.GetMetricAggregateRequest, opts ...grpc.CallOption) (*v2.GetMetricAggregateResponse, error) {
	if len(in.Metrics) == 0 {
		return &v2.GetMetricAggregateResponse{}, nil
	}
	metricId := in.Metrics[0]
	inv1 := &v1.GetMetricAggregateRequest{
		Dimensions:    toV1Dimensions(in.Dimensions),
		Metric:        metricId,
		AggregateType: v1.AggregateType(in.AggregateType),
		StartDate:     in.StartDate.AsTime().Unix(),
		EndDate:       in.EndDate.AsTime().Unix(),
		MaxItems:      in.MaxItems,
		TimeOrdering:  v1.TimeOrdering(in.TimeOrdering),
		StartingToken: in.StartingToken,
		IntervalMs:    in.IntervalMs,
	}
	res, err := b.v1Client.GetMetricAggregate(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	timestamps := make([]*timestamppb.Timestamp, len(res.Values))
	doubleValues := make([]float64, len(res.Values))
	for i := range res.Values {
		v := res.Values[i]
		if v == nil {
			continue
		}
		timestamps[i] = timestamppb.New(getTime(v.Timestamp))
		var value float64
		if v.Value != nil {
			value = v.Value.DoubleValue
		}
		doubleValues[i] = value
	}

	return &v2.GetMetricAggregateResponse{
		Frames: []*v2.Frame{
			{
				Metric:     metricId,
				Timestamps: timestamps,
				Fields: []*v2.Field{
					{
						Name:   "",
						Labels: nil,
						Config: nil,
						Values: doubleValues,
					},
				},
			},
		},
		NextToken: res.NextToken,
	}, nil
}

func getTime(timeInSeconds int64) time.Time {
	return time.Unix(timeInSeconds, 0)
}
