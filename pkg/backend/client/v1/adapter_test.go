package v1

import (
	"context"
	"testing"
	"time"

	v1 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v1"
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (v *v1Mock) GetMetricValue(ctx context.Context, in *v1.GetMetricValueRequest, opts ...grpc.CallOption) (*v1.GetMetricValueResponse, error) {
	args := v.Called(ctx, in)
	if v, ok := args.Get(0).(*v1.GetMetricValueResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
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

func TestAdapter_GetMetricValue(t *testing.T) {
	ts := time.Unix(1000, 0)

	req := &v3.GetMetricValueRequest{
		Dimensions: []*v3.Dimension{
			{
				Key:   "machine",
				Value: "m1",
			},
		},
		Metrics: []string{"foo"},
	}
	m := &v1Mock{}
	m.On("GetMetricValue", mock.Anything, &v1.GetMetricValueRequest{
		Dimensions: []*v1.Dimension{
			{Key: req.Dimensions[0].Key, Value: req.Dimensions[0].Value},
		},
		Metric: req.Metrics[0],
	}).Return(&v1.GetMetricValueResponse{
		Timestamp: ts.Unix(),
		Value: &v1.MetricValue{
			DoubleValue: 20,
		},
	}, nil)

	sut := &adapter{
		v1Client: m,
	}

	res, err := sut.GetMetricValue(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	m.AssertExpectations(t)
	expected := []*v3.GetMetricValueResponse_Frame{
		{
			Metric: req.Metrics[0],
			Fields: []*v3.SingleValueField{
				{
					Name:   "",
					Labels: nil,
					Config: nil,
					Value:  20,
				},
			},
			Timestamp: timestamppb.New(ts),
		},
	}
	assert.Equal(t, expected, res.Frames)
}

func TestAdapter_GetMetricHistory(t *testing.T) {
	req := &v3.GetMetricHistoryRequest{
		Dimensions: []*v3.Dimension{
			{
				Key:   "machine",
				Value: "m1",
			},
		},
		Metrics:       []string{"foo"},
		StartDate:     timestamppb.New(time.Unix(1000, 0)),
		EndDate:       timestamppb.New(time.Unix(2000, 0)),
		MaxItems:      30000,
		TimeOrdering:  v3.TimeOrdering_DESCENDING,
		StartingToken: "start-here",
	}

	m := &v1Mock{}
	m.On("GetMetricHistory", mock.Anything, &v1.GetMetricHistoryRequest{
		Dimensions: []*v1.Dimension{
			{Key: req.Dimensions[0].Key, Value: req.Dimensions[0].Value},
		},
		Metric:        req.Metrics[0],
		StartDate:     req.StartDate.Seconds,
		EndDate:       req.EndDate.Seconds,
		MaxItems:      req.MaxItems,
		TimeOrdering:  v1.TimeOrdering(req.TimeOrdering),
		StartingToken: req.StartingToken,
	}).Return(&v1.GetMetricHistoryResponse{
		Values: []*v1.MetricHistoryValue{
			{
				Timestamp: req.StartDate.Seconds,
				Value: &v1.MetricValue{
					DoubleValue: 2,
				},
			},
		},
		NextToken: "next-please",
	}, nil)

	sut := &adapter{
		v1Client: m,
	}

	res, err := sut.GetMetricHistory(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	m.AssertExpectations(t)
	expected := []*v3.Frame{
		{
			Metric: req.Metrics[0],
			Fields: []*v3.Field{
				{
					Name:   "",
					Labels: nil,
					Config: nil,
					Values: []float64{2},
				},
			},
			Timestamps: []*timestamppb.Timestamp{
				req.StartDate,
			},
		},
	}
	assert.Equal(t, expected, res.Frames)
	assert.Equal(t, "next-please", res.NextToken)
}

func TestAdapter_GetMetricAggregate(t *testing.T) {
	req := &v3.GetMetricAggregateRequest{
		Dimensions: []*v3.Dimension{
			{
				Key:   "machine",
				Value: "m1",
			},
		},
		Metrics:       []string{"foo"},
		StartDate:     timestamppb.New(time.Unix(1000, 0)),
		EndDate:       timestamppb.New(time.Unix(2000, 0)),
		MaxItems:      30000,
		TimeOrdering:  v3.TimeOrdering_DESCENDING,
		StartingToken: "start-here",
		IntervalMs:    999,
		Options:       map[string]string{aggregateTypeOptionKey: "3"},
	}

	m := &v1Mock{}
	m.On("GetMetricAggregate", mock.Anything, &v1.GetMetricAggregateRequest{
		Dimensions: []*v1.Dimension{
			{Key: req.Dimensions[0].Key, Value: req.Dimensions[0].Value},
		},
		Metric:        req.Metrics[0],
		AggregateType: v1.AggregateType_COUNT,
		StartDate:     req.StartDate.Seconds,
		EndDate:       req.EndDate.Seconds,
		MaxItems:      req.MaxItems,
		TimeOrdering:  v1.TimeOrdering(req.TimeOrdering),
		StartingToken: req.StartingToken,
		IntervalMs:    req.IntervalMs,
	}).Return(&v1.GetMetricAggregateResponse{
		Values: []*v1.MetricHistoryValue{
			{
				Timestamp: 1500,
				Value: &v1.MetricValue{
					DoubleValue: 2,
				},
			},
		},
		NextToken: "next-please",
	}, nil)

	sut := &adapter{
		v1Client: m,
	}

	res, err := sut.GetMetricAggregate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	m.AssertExpectations(t)
	expected := []*v3.Frame{
		{
			Metric: req.Metrics[0],
			Fields: []*v3.Field{
				{
					Name:   "",
					Labels: nil,
					Config: nil,
					Values: []float64{2},
				},
			},
			Timestamps: []*timestamppb.Timestamp{
				timestamppb.New(time.Unix(1500, 0)),
			},
		},
	}
	assert.Equal(t, expected, res.Frames)
	assert.Equal(t, "next-please", res.NextToken)
}
