package v1

import (
	"context"
	"strconv"
	"time"

	v1 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v1"
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type adapter struct {
	v1Client v1.GrafanaQueryAPIClient
}

// Gets the options for the specified query type
func (adapter *adapter) GetQueryOptions(ctx context.Context, in *v3.GetOptionsRequest, opts ...grpc.CallOption) (*v3.GetOptionsResponse, error) {
	if in.QueryType == v3.GetOptionsRequest_GetMetricAggregate {
		return &v3.GetOptionsResponse{
			Options: []*v3.Option{
				{
					Id:          aggregateTypeOptionID,
					Label:       "Aggregate",
					Type:        v3.Option_Enum,
					Description: "Selects the aggregate for metric values",
					EnumValues: []*v3.EnumValue{
						{Label: "Average", Description: "Average value aggregate", Id: strconv.Itoa(int(v1.AggregateType_AVERAGE))},
						{Label: "Min", Description: "Min value aggregate", Id: strconv.Itoa(int(v1.AggregateType_MIN))},
						{Label: "Max", Description: "Max value aggregate", Id: strconv.Itoa(int(v1.AggregateType_MAX))},
						{Label: "Count", Description: "Count value aggregate", Id: strconv.Itoa(int(v1.AggregateType_COUNT))},
					}},
			},
		}, nil
	}
	return &v3.GetOptionsResponse{}, nil
}

func (b *adapter) ListDimensionKeys(ctx context.Context, in *v3.ListDimensionKeysRequest, opts ...grpc.CallOption) (*v3.ListDimensionKeysResponse, error) {
	inv1 := &v1.ListDimensionKeysRequest{
		Filter: in.Filter,
	}
	res, err := b.v1Client.ListDimensionKeys(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	r := make([]*v3.ListDimensionKeysResponse_Result, len(res.Results))
	for i := range res.Results {
		r[i] = &v3.ListDimensionKeysResponse_Result{
			Key:         res.Results[i].Key,
			Description: res.Results[i].Description,
		}
	}
	return &v3.ListDimensionKeysResponse{
		Results: r,
	}, nil
}

func (b *adapter) ListDimensionValues(ctx context.Context, in *v3.ListDimensionValuesRequest, opts ...grpc.CallOption) (*v3.ListDimensionValuesResponse, error) {
	inv1 := &v1.ListDimensionValuesRequest{
		DimensionKey: in.DimensionKey,
		Filter:       in.Filter,
	}
	res, err := b.v1Client.ListDimensionValues(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	r := make([]*v3.ListDimensionValuesResponse_Result, len(res.Results))
	for i := range res.Results {
		r[i] = &v3.ListDimensionValuesResponse_Result{
			Value:       res.Results[i].Value,
			Description: res.Results[i].Description,
		}
	}
	return &v3.ListDimensionValuesResponse{
		Results: r,
	}, nil
}

func (b *adapter) ListMetrics(ctx context.Context, in *v3.ListMetricsRequest, opts ...grpc.CallOption) (*v3.ListMetricsResponse, error) {
	inv1 := &v1.ListMetricsRequest{
		Dimensions: toV1Dimensions(in.Dimensions),
		Filter:     in.Filter,
	}
	res, err := b.v1Client.ListMetrics(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	r := make([]*v3.ListMetricsResponse_Metric, len(res.Metrics))
	for i := range res.Metrics {
		r[i] = &v3.ListMetricsResponse_Metric{
			Name:        res.Metrics[i].Name,
			Description: res.Metrics[i].Description,
		}
	}
	return &v3.ListMetricsResponse{
		Metrics: r,
	}, nil
}

func (b *adapter) GetMetricValue(ctx context.Context, in *v3.GetMetricValueRequest, opts ...grpc.CallOption) (*v3.GetMetricValueResponse, error) {
	if len(in.Metrics) == 0 {
		return &v3.GetMetricValueResponse{}, nil
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
	return &v3.GetMetricValueResponse{
		Frames: []*v3.GetMetricValueResponse_Frame{
			{
				Metric:    metricId,
				Timestamp: timestamppb.New(getTime(res.Timestamp)),
				Fields: []*v3.SingleValueField{
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

func toV1Dimensions(dims []*v3.Dimension) []*v1.Dimension {
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

func (b *adapter) GetMetricHistory(ctx context.Context, in *v3.GetMetricHistoryRequest, opts ...grpc.CallOption) (*v3.GetMetricHistoryResponse, error) {
	if len(in.Metrics) == 0 {
		return &v3.GetMetricHistoryResponse{}, nil
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
	return &v3.GetMetricHistoryResponse{
		Frames: []*v3.Frame{
			{
				Metric:     metricId,
				Timestamps: timestamps,
				Fields: []*v3.Field{
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

const aggregateTypeOptionID = "0"

func (b *adapter) GetMetricAggregate(ctx context.Context, in *v3.GetMetricAggregateRequest, opts ...grpc.CallOption) (*v3.GetMetricAggregateResponse, error) {
	if len(in.Metrics) == 0 {
		return &v3.GetMetricAggregateResponse{}, nil
	}
	metricId := in.Metrics[0]

	var aggregateType v1.AggregateType
	switch in.GetOptions()[aggregateTypeOptionID] {
	case "0":
		aggregateType = v1.AggregateType_AVERAGE
	case "1":
		aggregateType = v1.AggregateType_MIN
	case "2":
		aggregateType = v1.AggregateType_MAX
	case "3":
		aggregateType = v1.AggregateType_COUNT
	default:
		aggregateType = v1.AggregateType_AVERAGE
	}

	inv1 := &v1.GetMetricAggregateRequest{
		Dimensions:    toV1Dimensions(in.Dimensions),
		Metric:        metricId,
		AggregateType: aggregateType,
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

	return &v3.GetMetricAggregateResponse{
		Frames: []*v3.Frame{
			{
				Metric:     metricId,
				Timestamps: timestamps,
				Fields: []*v3.Field{
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
