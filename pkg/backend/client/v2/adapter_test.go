package v2

import (
	"context"
	"testing"
	"time"

	v2 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type v1Mock struct {
	mock.Mock
}

func (v v1Mock) ListDimensionKeys(ctx context.Context, in *v2.ListDimensionKeysRequest, opts ...grpc.CallOption) (*v2.ListDimensionKeysResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (v v1Mock) ListDimensionValues(ctx context.Context, in *v2.ListDimensionValuesRequest, opts ...grpc.CallOption) (*v2.ListDimensionValuesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (v v1Mock) ListMetrics(ctx context.Context, in *v2.ListMetricsRequest, opts ...grpc.CallOption) (*v2.ListMetricsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (v *v1Mock) GetMetricValue(ctx context.Context, in *v2.GetMetricValueRequest, opts ...grpc.CallOption) (*v2.GetMetricValueResponse, error) {
	args := v.Called(ctx, in)
	if v, ok := args.Get(0).(*v2.GetMetricValueResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (v *v1Mock) GetMetricHistory(ctx context.Context, in *v2.GetMetricHistoryRequest, opts ...grpc.CallOption) (*v2.GetMetricHistoryResponse, error) {
	args := v.Called(ctx, in)
	if v, ok := args.Get(0).(*v2.GetMetricHistoryResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (v *v1Mock) GetMetricAggregate(ctx context.Context, in *v2.GetMetricAggregateRequest, opts ...grpc.CallOption) (*v2.GetMetricAggregateResponse, error) {
	args := v.Called(ctx, in)
	if v, ok := args.Get(0).(*v2.GetMetricAggregateResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}
func mustParseTime(s string) *timestamppb.Timestamp {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return timestamppb.New(t)
}

func TestAdapter_GetMetricValue(t *testing.T) {
	ts := mustParseTime("2022-07-20T12:26:06Z")

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
	v2Response := &v2.GetMetricValueResponse{
		Frames: []*v2.GetMetricValueResponse_Frame{
			{
				Metric:    "foo",
				Timestamp: ts,
				Fields: []*v2.SingleValueField{
					{
						Name:  "value",
						Value: 12.42,
						Config: &v2.Config{
							Unit: "mm",
							Mappings: []*v2.ValueMapping{
								{From: 1, To: 2, Value: "FOO", Text: "BAR", Color: "yellow"},
							},
						},
						Labels: []*v2.Label{{Key: "foo", Value: "bar"}}},
				},
			},
		},
	}
	m.On("GetMetricValue", mock.Anything, &v2.GetMetricValueRequest{
		Dimensions: []*v2.Dimension{
			{Key: req.Dimensions[0].Key, Value: req.Dimensions[0].Value},
		},
		Metrics: req.Metrics,
	}).Return(v2Response, nil)

	sut := &adapter{
		v2Client: m,
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
					Name:   "value",
					Labels: []*v3.Label{{Key: "foo", Value: "bar"}},
					Value:  12.42,
					Config: &v3.Config{
						Unit: "mm",
						Mappings: []*v3.ValueMapping{
							{From: 1, To: 2, Value: "FOO", Text: "BAR", Color: "yellow"},
						},
					},
				},
			},
			Timestamp: ts,
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
	ts := mustParseTime("2022-07-20T12:38:01Z")
	v2Frame := &v2.Frame{
		Metric:     "foo",
		Timestamps: []*timestamppb.Timestamp{ts},
		Fields: []*v2.Field{
			{
				Name:   "value",
				Values: []float64{1.42},
				Config: &v2.Config{
					Unit: "mm",
					Mappings: []*v2.ValueMapping{
						{From: 1, To: 2, Value: "FOO", Text: "BAR", Color: "yellow"},
					},
				},
				Labels: []*v2.Label{{Key: "foo", Value: "bar"}},
			},
		},
		Meta: &v2.FrameMeta{
			Type:                   v2.FrameMeta_FrameTypeDirectoryListing,
			PreferredVisualization: v2.FrameMeta_VisTypeLogs,
			ExecutedQueryString:    "foo bar baz",
			Notices: []*v2.FrameMeta_Notice{
				{Severity: v2.FrameMeta_Notice_NoticeSeverityWarning, Text: "This is a notice", Link: "https://foo.bar", Inspect: v2.FrameMeta_Notice_InspectTypeError},
			},
		},
	}
	m.On("GetMetricHistory", mock.Anything, &v2.GetMetricHistoryRequest{
		Dimensions: []*v2.Dimension{
			{Key: req.Dimensions[0].Key, Value: req.Dimensions[0].Value},
		},
		Metrics:       req.Metrics,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		MaxItems:      req.MaxItems,
		TimeOrdering:  v2.TimeOrdering(req.TimeOrdering),
		StartingToken: req.StartingToken,
	}).Return(&v2.GetMetricHistoryResponse{
		Frames:    []*v2.Frame{v2Frame},
		NextToken: "next-please",
	}, nil)

	sut := &adapter{
		v2Client: m,
	}

	res, err := sut.GetMetricHistory(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	m.AssertExpectations(t)
	expected := []*v3.Frame{
		{
			Metric: v2Frame.Metric,
			Fields: []*v3.Field{
				{

					Name:   "value",
					Values: []float64{1.42},
					Config: &v3.Config{
						Unit: "mm",
						Mappings: []*v3.ValueMapping{
							{From: 1, To: 2, Value: "FOO", Text: "BAR", Color: "yellow"},
						},
					},
					Labels: []*v3.Label{{Key: "foo", Value: "bar"}},
				},
			},
			Timestamps: []*timestamppb.Timestamp{
				ts,
			},
			Meta: &v3.FrameMeta{
				Type:                   v3.FrameMeta_FrameTypeDirectoryListing,
				PreferredVisualization: v3.FrameMeta_VisTypeLogs,
				ExecutedQueryString:    "foo bar baz",
				Notices: []*v3.FrameMeta_Notice{
					{Severity: v3.FrameMeta_Notice_NoticeSeverityWarning, Text: "This is a notice", Link: "https://foo.bar", Inspect: v3.FrameMeta_Notice_InspectTypeError},
				},
			},
		},
	}
	assert.Equal(t, expected, res.Frames)
	assert.Equal(t, "next-please", res.NextToken)
}

func TestAdapter_GetMetricAggregate(t *testing.T) {

	ts := mustParseTime("2022-07-20T12:38:01Z")
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

	v2Frame := &v2.Frame{
		Metric:     "foo",
		Timestamps: []*timestamppb.Timestamp{ts},
		Fields: []*v2.Field{
			{
				Name:   "value",
				Values: []float64{1.42},
				Config: &v2.Config{
					Unit: "mm",
					Mappings: []*v2.ValueMapping{
						{From: 1, To: 2, Value: "FOO", Text: "BAR", Color: "yellow"},
					},
				},
				Labels: []*v2.Label{{Key: "foo", Value: "bar"}},
			},
		},
		Meta: &v2.FrameMeta{
			Type:                   v2.FrameMeta_FrameTypeDirectoryListing,
			PreferredVisualization: v2.FrameMeta_VisTypeLogs,
			ExecutedQueryString:    "foo bar baz",
			Notices: []*v2.FrameMeta_Notice{
				{Severity: v2.FrameMeta_Notice_NoticeSeverityWarning, Text: "This is a notice", Link: "https://foo.bar", Inspect: v2.FrameMeta_Notice_InspectTypeError},
			},
		},
	}
	m := &v1Mock{}
	m.On("GetMetricAggregate", mock.Anything, &v2.GetMetricAggregateRequest{
		Dimensions: []*v2.Dimension{
			{Key: req.Dimensions[0].Key, Value: req.Dimensions[0].Value},
		},
		Metrics:       req.Metrics,
		AggregateType: v2.AggregateType_COUNT,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		MaxItems:      req.MaxItems,
		TimeOrdering:  v2.TimeOrdering(req.TimeOrdering),
		StartingToken: req.StartingToken,
		IntervalMs:    req.IntervalMs,
	}).Return(&v2.GetMetricAggregateResponse{
		Frames:    []*v2.Frame{v2Frame},
		NextToken: "next-please",
	}, nil)

	sut := &adapter{
		v2Client: m,
	}

	res, err := sut.GetMetricAggregate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	m.AssertExpectations(t)

	expected := []*v3.Frame{
		{
			Metric: v2Frame.Metric,
			Fields: []*v3.Field{
				{

					Name:   "value",
					Values: []float64{1.42},
					Config: &v3.Config{
						Unit: "mm",
						Mappings: []*v3.ValueMapping{
							{From: 1, To: 2, Value: "FOO", Text: "BAR", Color: "yellow"},
						},
					},
					Labels: []*v3.Label{{Key: "foo", Value: "bar"}},
				},
			},
			Timestamps: []*timestamppb.Timestamp{
				ts,
			},
			Meta: &v3.FrameMeta{
				Type:                   v3.FrameMeta_FrameTypeDirectoryListing,
				PreferredVisualization: v3.FrameMeta_VisTypeLogs,
				ExecutedQueryString:    "foo bar baz",
				Notices: []*v3.FrameMeta_Notice{
					{Severity: v3.FrameMeta_Notice_NoticeSeverityWarning, Text: "This is a notice", Link: "https://foo.bar", Inspect: v3.FrameMeta_Notice_InspectTypeError},
				},
			},
		},
	}
	assert.Equal(t, expected, res.Frames)
	assert.Equal(t, "next-please", res.NextToken)
}
