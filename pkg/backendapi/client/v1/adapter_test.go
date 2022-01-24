package v1

import (
	v1 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v1"
	v2 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"testing"
)

type v1Mock struct {
	mock.Mock
}

func (v v1Mock) ListDimensionKeys(ctx context.Context, in *v1.ListDimensionKeysRequest, opts ...grpc.CallOption) (*v1.ListDimensionKeysResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (v v1Mock) ListDimensionValues(ctx context.Context, in *v1.ListDimensionValuesRequest, opts ...grpc.CallOption) (*v1.ListDimensionValuesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (v v1Mock) ListMetrics(ctx context.Context, in *v1.ListMetricsRequest, opts ...grpc.CallOption) (*v1.ListMetricsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (v v1Mock) GetMetricValue(ctx context.Context, in *v1.GetMetricValueRequest, opts ...grpc.CallOption) (*v1.GetMetricValueResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (v *v1Mock) GetMetricHistory(ctx context.Context, in *v1.GetMetricHistoryRequest, opts ...grpc.CallOption) (*v1.GetMetricHistoryResponse, error) {
	args := v.Called(ctx, in)
	if v, ok := args.Get(0).(*v1.GetMetricHistoryResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (v *v1Mock) GetMetricAggregate(ctx context.Context, in *v1.GetMetricAggregateRequest, opts ...grpc.CallOption) (*v1.GetMetricAggregateResponse, error) {
	args := v.Called(ctx, in)
	if v, ok := args.Get(0).(*v1.GetMetricAggregateResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestAdapter_GetMetricHistory(t *testing.T) {
	req := &v2.GetMetricHistoryRequest{
		Dimensions: []*v2.Dimension{
			{
				Key:   "machine",
				Value: "m1",
			},
		},
		Metric:        "foo",
		StartDate:     10000,
		EndDate:       20000,
		MaxItems:      30000,
		TimeOrdering:  v2.TimeOrdering_DESCENDING,
		StartingToken: "start-here",
	}

	m := &v1Mock{}
	m.On("GetMetricHistory", mock.Anything, &v1.GetMetricHistoryRequest{
		Dimensions: []*v1.Dimension{
			{Key: req.Dimensions[0].Key, Value: req.Dimensions[0].Value},
		},
		Metric:        req.Metric,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		MaxItems:      req.MaxItems,
		TimeOrdering:  v1.TimeOrdering(req.TimeOrdering),
		StartingToken: req.StartingToken,
	}).Return(&v1.GetMetricHistoryResponse{
		Values: []*v1.MetricHistoryValue{
			{
				Timestamp: 1,
				Value: &v1.MetricValue{
					DoubleValue: 2,
				},
			},
		},
		NextToken: "next-please",
	}, nil)

	sut := &adapter{
		adaptee: m,
	}

	res, err := sut.GetMetricHistory(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	m.AssertExpectations(t)
	expected := []*v2.GetMetricHistoryResponse_Data{
		{
			Metric: req.Metric,
			Series: []*v2.GetMetricHistoryResponse_Data_TimeSeries{
				{
					Timestamp: 1,
					Values: []*v2.GetMetricHistoryResponse_Data_TimeSeries_MetricValue{
						{
							DoubleValue: 2,
							Id:          "",
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expected, res.Data)
	assert.Equal(t, "next-please", res.NextToken)
}

func TestAdapter_GetMetricAggregate(t *testing.T) {
	req := &v2.GetMetricAggregateRequest{
		Dimensions: []*v2.Dimension{
			{
				Key:   "machine",
				Value: "m1",
			},
		},
		Metric:        "foo",
		AggregateType: v2.AggregateType_COUNT,
		StartDate:     10000,
		EndDate:       20000,
		MaxItems:      30000,
		TimeOrdering:  v2.TimeOrdering_DESCENDING,
		StartingToken: "start-here",
		IntervalMs:    999,
	}

	m := &v1Mock{}
	m.On("GetMetricAggregate", mock.Anything, &v1.GetMetricAggregateRequest{
		Dimensions: []*v1.Dimension{
			{Key: req.Dimensions[0].Key, Value: req.Dimensions[0].Value},
		},
		Metric:        req.Metric,
		AggregateType: v1.AggregateType(req.AggregateType),
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		MaxItems:      req.MaxItems,
		TimeOrdering:  v1.TimeOrdering(req.TimeOrdering),
		StartingToken: req.StartingToken,
		IntervalMs:    req.IntervalMs,
	}).Return(&v1.GetMetricAggregateResponse{
		Values: []*v1.MetricHistoryValue{
			{
				Timestamp: 1,
				Value: &v1.MetricValue{
					DoubleValue: 2,
				},
			},
		},
		NextToken: "next-please",
	}, nil)

	sut := &adapter{
		adaptee: m,
	}

	res, err := sut.GetMetricAggregate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	m.AssertExpectations(t)
	expected := []*v2.GetMetricAggregateResponse_Data{
		{
			Metric: req.Metric,
			Series: []*v2.GetMetricAggregateResponse_Data_TimeSeries{
				{
					Timestamp: 1,
					Values: []*v2.GetMetricAggregateResponse_Data_TimeSeries_MetricValue{
						{
							DoubleValue:   2,
							AggregateType: req.AggregateType,
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expected, res.Data)
	assert.Equal(t, "next-please", res.NextToken)
}
