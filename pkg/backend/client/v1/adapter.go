package v1

import (
	"context"
	"strconv"
	"time"

	v1 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v1"
	v4 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v4"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type adapter struct {
	v1Client v1.GrafanaQueryAPIClient
}

func (adapter *adapter) GetStreamingQueryConfiguration(ctx context.Context, in *v4.GetStreamingQueryConfigurationRequest, opts ...grpc.CallOption) (*v4.GetStreamingQueryConfigurationResponse, error) {
	return nil, nil
}

// Gets the options for the specified query type
func (adapter *adapter) GetQueryOptions(ctx context.Context, in *v4.GetOptionsRequest, opts ...grpc.CallOption) (*v4.GetOptionsResponse, error) {
	if in.QueryType == v4.GetOptionsRequest_GetMetricAggregate {
		return &v4.GetOptionsResponse{
			Options: []*v4.Option{
				{
					Id:          aggregateTypeOptionID,
					Label:       "Aggregate",
					Type:        v4.Option_Enum,
					Description: "Selects the aggregate for metric values",
					EnumValues: []*v4.EnumValue{
						{Label: "Average", Description: "Average value aggregate", Id: strconv.Itoa(int(v1.AggregateType_AVERAGE))},
						{Label: "Min", Description: "Min value aggregate", Id: strconv.Itoa(int(v1.AggregateType_MIN))},
						{Label: "Max", Description: "Max value aggregate", Id: strconv.Itoa(int(v1.AggregateType_MAX))},
						{Label: "Count", Description: "Count value aggregate", Id: strconv.Itoa(int(v1.AggregateType_COUNT))},
					}},
			},
		}, nil
	}
	return &v4.GetOptionsResponse{}, nil
}

func (b *adapter) ListDimensionKeys(ctx context.Context, in *v4.ListDimensionKeysRequest, opts ...grpc.CallOption) (*v4.ListDimensionKeysResponse, error) {
	inv1 := &v1.ListDimensionKeysRequest{
		Filter: in.Filter,
	}
	res, err := b.v1Client.ListDimensionKeys(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	r := make([]*v4.ListDimensionKeysResponse_Result, len(res.Results))
	for i := range res.Results {
		r[i] = &v4.ListDimensionKeysResponse_Result{
			Key:         res.Results[i].Key,
			Description: res.Results[i].Description,
		}
	}
	return &v4.ListDimensionKeysResponse{
		Results: r,
	}, nil
}

func (b *adapter) ListDimensionValues(ctx context.Context, in *v4.ListDimensionValuesRequest, opts ...grpc.CallOption) (*v4.ListDimensionValuesResponse, error) {
	inv1 := &v1.ListDimensionValuesRequest{
		DimensionKey: in.DimensionKey,
		Filter:       in.Filter,
	}
	res, err := b.v1Client.ListDimensionValues(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	r := make([]*v4.ListDimensionValuesResponse_Result, len(res.Results))
	for i := range res.Results {
		r[i] = &v4.ListDimensionValuesResponse_Result{
			Value:       res.Results[i].Value,
			Description: res.Results[i].Description,
		}
	}
	return &v4.ListDimensionValuesResponse{
		Results: r,
	}, nil
}

func (b *adapter) ListMetrics(ctx context.Context, in *v4.ListMetricsRequest, opts ...grpc.CallOption) (*v4.ListMetricsResponse, error) {
	inv1 := &v1.ListMetricsRequest{
		Dimensions: toV1Dimensions(in.Dimensions),
		Filter:     in.Filter,
	}
	res, err := b.v1Client.ListMetrics(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	r := make([]*v4.ListMetricsResponse_Metric, len(res.Metrics))
	for i := range res.Metrics {
		r[i] = &v4.ListMetricsResponse_Metric{
			Name:        res.Metrics[i].Name,
			Description: res.Metrics[i].Description,
		}
	}
	return &v4.ListMetricsResponse{
		Metrics: r,
	}, nil
}

func (b *adapter) GetMetricValue(ctx context.Context, in *v4.GetMetricValueRequest, opts ...grpc.CallOption) (*v4.GetMetricValueResponse, error) {
	if len(in.Metrics) == 0 {
		return &v4.GetMetricValueResponse{}, nil
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
	return &v4.GetMetricValueResponse{
		Frames: []*v4.GetMetricValueResponse_Frame{
			{
				Metric:    metricId,
				Timestamp: timestamppb.New(getTime(res.Timestamp)),
				Fields: []*v4.SingleValueField{
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

func toV1Dimensions(dims []*v4.Dimension) []*v1.Dimension {
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

func (b *adapter) GetMetricHistory(ctx context.Context, in *v4.GetMetricHistoryRequest, opts ...grpc.CallOption) (*v4.GetMetricHistoryResponse, error) {
	if len(in.Metrics) == 0 {
		return &v4.GetMetricHistoryResponse{}, nil
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
	return &v4.GetMetricHistoryResponse{
		Frames: []*v4.Frame{
			{
				Metric:     metricId,
				Timestamps: timestamps,
				Fields: []*v4.Field{
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

func (b *adapter) GetMetricAggregate(ctx context.Context, in *v4.GetMetricAggregateRequest, opts ...grpc.CallOption) (*v4.GetMetricAggregateResponse, error) {
	if len(in.Metrics) == 0 {
		return &v4.GetMetricAggregateResponse{}, nil
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

	return &v4.GetMetricAggregateResponse{
		Frames: []*v4.Frame{
			{
				Metric:     metricId,
				Timestamps: timestamps,
				Fields: []*v4.Field{
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
