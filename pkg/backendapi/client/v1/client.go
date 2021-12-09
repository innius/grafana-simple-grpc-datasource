package v1

import (
	v1 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v1"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"context"
	"google.golang.org/grpc"
)

type adapter struct {
	adaptee v1.GrafanaQueryAPIClient
}

func NewClient(conn *grpc.ClientConn) (v2.GrafanaQueryAPIClient, error) {
	return &adapter{adaptee: v1.NewGrafanaQueryAPIClient(conn)}, nil
}

func (b *adapter) ListDimensionKeys(ctx context.Context, in *v2.ListDimensionKeysRequest, opts ...grpc.CallOption) (*v2.ListDimensionKeysResponse, error) {
	inv1 := &v1.ListDimensionKeysRequest{
		Filter: in.Filter,
	}
	res, err := b.adaptee.ListDimensionKeys(ctx, inv1, opts...)
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
	res, err := b.adaptee.ListDimensionValues(ctx, inv1, opts...)
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
	res, err := b.adaptee.ListMetrics(ctx, inv1, opts...)
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
	inv1 := &v1.GetMetricValueRequest{
		Dimensions: toV1Dimensions(in.Dimensions),
		Metric:     in.Metric,
	}
	res, err := b.adaptee.GetMetricValue(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	var values []*v2.MetricValue
	if res.Value != nil {
		values = append(values, &v2.MetricValue{
			DoubleValue: res.Value.DoubleValue,
		})
	}
	return &v2.GetMetricValueResponse{
		Timestamp: res.Timestamp,
		Values:    values,
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

func toV1TimeOrdering(in v2.TimeOrdering) v1.TimeOrdering {
	if in == v2.TimeOrdering_ASCENDING {
		return v1.TimeOrdering_ASCENDING
	}
	return v1.TimeOrdering_DESCENDING
}

func (b *adapter) GetMetricHistory(ctx context.Context, in *v2.GetMetricHistoryRequest, opts ...grpc.CallOption) (*v2.GetMetricHistoryResponse, error) {
	inv1 := &v1.GetMetricHistoryRequest{
		Dimensions:    toV1Dimensions(in.Dimensions),
		Metric:        in.Metric,
		StartDate:     in.StartDate,
		EndDate:       in.EndDate,
		MaxItems:      in.MaxItems,
		TimeOrdering:  toV1TimeOrdering(in.TimeOrdering),
		StartingToken: in.StartingToken,
	}
	res, err := b.adaptee.GetMetricHistory(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	values := make([]*v2.MetricHistoryValue, len(res.Values))
	for i := range res.Values {
		v := res.Values[i]
		if v == nil {
			continue
		}
		values[i] = &v2.MetricHistoryValue{
			Timestamp: v.Timestamp,
			Values: []*v2.MetricValue{
				{
					DoubleValue: v.Value.DoubleValue,
				},
			},
		}
	}
	return &v2.GetMetricHistoryResponse{
		Values:    values,
		NextToken: res.NextToken,
	}, nil
}

func (b *adapter) GetMetricAggregate(ctx context.Context, in *v2.GetMetricAggregateRequest, opts ...grpc.CallOption) (*v2.GetMetricAggregateResponse, error) {
	inv1 := &v1.GetMetricAggregateRequest{
		Dimensions:    toV1Dimensions(in.Dimensions),
		Metric:        in.Metric,
		StartDate:     in.StartDate,
		EndDate:       in.EndDate,
		MaxItems:      in.MaxItems,
		TimeOrdering:  toV1TimeOrdering(in.TimeOrdering),
		StartingToken: in.StartingToken,
	}
	res, err := b.adaptee.GetMetricAggregate(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	values := make([]*v2.MetricHistoryValue, len(res.Values))
	for i := range res.Values {
		v := res.Values[i]
		if v == nil {
			continue
		}
		values[i] = &v2.MetricHistoryValue{
			Timestamp: v.Timestamp,
			Values: []*v2.MetricValue{
				{
					DoubleValue: v.Value.DoubleValue,
				},
			},
		}
	}
	return &v2.GetMetricAggregateResponse{
		Values:    values,
		NextToken: res.NextToken,
	}, nil
}

