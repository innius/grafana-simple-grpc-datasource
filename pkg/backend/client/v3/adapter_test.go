package v3

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	v4 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v4"
)

type adapteeMock struct {
	mock.Mock
}

// Gets the options for the specified query type
func (v *adapteeMock) GetQueryOptions(ctx context.Context, in *v3.GetOptionsRequest, opts ...grpc.CallOption) (*v3.GetOptionsResponse, error) {
	args := v.Called(ctx, in)
	if v, ok := args.Get(0).(*v3.GetOptionsResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (v adapteeMock) ListDimensionKeys(ctx context.Context, in *v3.ListDimensionKeysRequest, opts ...grpc.CallOption) (*v3.ListDimensionKeysResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (v adapteeMock) ListDimensionValues(ctx context.Context, in *v3.ListDimensionValuesRequest, opts ...grpc.CallOption) (*v3.ListDimensionValuesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (v adapteeMock) ListMetrics(ctx context.Context, in *v3.ListMetricsRequest, opts ...grpc.CallOption) (*v3.ListMetricsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (v *adapteeMock) GetMetricValue(ctx context.Context, in *v3.GetMetricValueRequest, opts ...grpc.CallOption) (*v3.GetMetricValueResponse, error) {
	args := v.Called(ctx, in)
	if v, ok := args.Get(0).(*v3.GetMetricValueResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (v *adapteeMock) GetMetricHistory(ctx context.Context, in *v3.GetMetricHistoryRequest, opts ...grpc.CallOption) (*v3.GetMetricHistoryResponse, error) {
	args := v.Called(ctx, in)
	if v, ok := args.Get(0).(*v3.GetMetricHistoryResponse); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (v *adapteeMock) GetMetricAggregate(ctx context.Context, in *v3.GetMetricAggregateRequest, opts ...grpc.CallOption) (*v3.GetMetricAggregateResponse, error) {
	args := v.Called(ctx, in)
	if v, ok := args.Get(0).(*v3.GetMetricAggregateResponse); ok {
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

	req := &v4.GetMetricValueRequest{
		Dimensions: []*v4.Dimension{
			{
				Key:   "machine",
				Value: "m1",
			},
		},
		Metrics: []string{"foo"},
	}
	m := &adapteeMock{}
	v3Response := &v3.GetMetricValueResponse{
		Frames: []*v3.GetMetricValueResponse_Frame{
			{
				Metric:    "foo",
				Timestamp: ts,
				Fields: []*v3.SingleValueField{
					{
						Name:  "value",
						Value: 12.42,
						Config: &v3.Config{
							Unit: "mm",
							Mappings: []*v3.ValueMapping{
								{From: 1, To: 2, Value: "FOO", Text: "BAR", Color: "yellow"},
							},
						},
						Labels: []*v3.Label{{Key: "foo", Value: "bar"}}},
				},
			},
		},
	}
	m.On("GetMetricValue", mock.Anything, &v3.GetMetricValueRequest{
		Dimensions: []*v3.Dimension{
			{Key: req.Dimensions[0].Key, Value: req.Dimensions[0].Value},
		},
		Metrics: req.Metrics,
	}).Return(v3Response, nil)

	sut := &adapter{
		adaptee: m,
	}

	res, err := sut.GetMetricValue(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	m.AssertExpectations(t)
	expected := []*v4.GetMetricValueResponse_Frame{
		{
			Metric: req.Metrics[0],
			Fields: []*v4.SingleValueField{
				{
					Name:   "value",
					Labels: []*v4.Label{{Key: "foo", Value: "bar"}},
					Value:  12.42,
					Config: &v4.Config{
						Unit: "mm",
						Mappings: []*v4.ValueMapping{
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
	req := &v4.GetMetricHistoryRequest{
		Dimensions: []*v4.Dimension{
			{
				Key:   "machine",
				Value: "m1",
			},
		},
		Metrics:       []string{"foo"},
		StartDate:     timestamppb.New(time.Unix(1000, 0)),
		EndDate:       timestamppb.New(time.Unix(2000, 0)),
		MaxItems:      30000,
		TimeOrdering:  v4.TimeOrdering_DESCENDING,
		StartingToken: "start-here",
	}

	m := &adapteeMock{}
	ts := mustParseTime("2022-07-20T12:38:01Z")
	v3Frame := &v3.Frame{
		Metric:     "foo",
		Timestamps: []*timestamppb.Timestamp{ts},
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
		Meta: &v3.FrameMeta{
			Type:                   v3.FrameMeta_FrameTypeDirectoryListing,
			PreferredVisualization: v3.FrameMeta_VisTypeLogs,
			ExecutedQueryString:    "foo bar baz",
			Notices: []*v3.FrameMeta_Notice{
				{Severity: v3.FrameMeta_Notice_NoticeSeverityWarning, Text: "This is a notice", Link: "https://foo.bar", Inspect: v3.FrameMeta_Notice_InspectTypeError},
			},
		},
	}
	m.On("GetMetricHistory", mock.Anything, &v3.GetMetricHistoryRequest{
		Dimensions: []*v3.Dimension{
			{Key: req.Dimensions[0].Key, Value: req.Dimensions[0].Value},
		},
		Metrics:       req.Metrics,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		MaxItems:      req.MaxItems,
		TimeOrdering:  v3.TimeOrdering(req.TimeOrdering),
		StartingToken: req.StartingToken,
	}).Return(&v3.GetMetricHistoryResponse{
		Frames:    []*v3.Frame{v3Frame},
		NextToken: "next-please",
	}, nil)

	sut := &adapter{
		adaptee: m,
	}

	res, err := sut.GetMetricHistory(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	m.AssertExpectations(t)
	expected := []*v4.Frame{
		{
			Metric: v3Frame.Metric,
			Fields: []*v4.Field{
				{

					Name:   "value",
					Values: []float64{1.42},
					Config: &v4.Config{
						Unit: "mm",
						Mappings: []*v4.ValueMapping{
							{From: 1, To: 2, Value: "FOO", Text: "BAR", Color: "yellow"},
						},
					},
					Labels: []*v4.Label{{Key: "foo", Value: "bar"}},
				},
			},
			Timestamps: []*timestamppb.Timestamp{
				ts,
			},
			Meta: &v4.FrameMeta{
				Type:                   v4.FrameMeta_FrameTypeDirectoryListing,
				PreferredVisualization: v4.FrameMeta_VisTypeLogs,
				ExecutedQueryString:    "foo bar baz",
				Notices: []*v4.FrameMeta_Notice{
					{Severity: v4.FrameMeta_Notice_NoticeSeverityWarning, Text: "This is a notice", Link: "https://foo.bar", Inspect: v4.FrameMeta_Notice_InspectTypeError},
				},
			},
		},
	}
	assert.Equal(t, expected, res.Frames)
	assert.Equal(t, "next-please", res.NextToken)
}

func TestAdapter_GetMetricAggregate(t *testing.T) {

	ts := mustParseTime("2022-07-20T12:38:01Z")
	req := &v4.GetMetricAggregateRequest{
		Dimensions: []*v4.Dimension{
			{
				Key:   "machine",
				Value: "m1",
			},
		},
		Metrics:       []string{"foo"},
		StartDate:     timestamppb.New(time.Unix(1000, 0)),
		EndDate:       timestamppb.New(time.Unix(2000, 0)),
		MaxItems:      30000,
		TimeOrdering:  v4.TimeOrdering_DESCENDING,
		StartingToken: "start-here",
		IntervalMs:    999,
		Options:       map[string]string{"foo": "bar"},
	}

	v3Frame := &v3.Frame{
		Metric:     "foo",
		Timestamps: []*timestamppb.Timestamp{ts},
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
		Meta: &v3.FrameMeta{
			Type:                   v3.FrameMeta_FrameTypeDirectoryListing,
			PreferredVisualization: v3.FrameMeta_VisTypeLogs,
			ExecutedQueryString:    "foo bar baz",
			Notices: []*v3.FrameMeta_Notice{
				{Severity: v3.FrameMeta_Notice_NoticeSeverityWarning, Text: "This is a notice", Link: "https://foo.bar", Inspect: v3.FrameMeta_Notice_InspectTypeError},
			},
		},
	}
	m := &adapteeMock{}
	m.On("GetMetricAggregate", mock.Anything, &v3.GetMetricAggregateRequest{
		Dimensions: []*v3.Dimension{
			{Key: req.Dimensions[0].Key, Value: req.Dimensions[0].Value},
		},
		Metrics:       req.Metrics,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		MaxItems:      req.MaxItems,
		TimeOrdering:  v3.TimeOrdering(req.TimeOrdering),
		StartingToken: req.StartingToken,
		IntervalMs:    req.IntervalMs,
		Options:       req.Options,
	}).Return(&v3.GetMetricAggregateResponse{
		Frames:    []*v3.Frame{v3Frame},
		NextToken: "next-please",
	}, nil)

	sut := &adapter{
		adaptee: m,
	}

	res, err := sut.GetMetricAggregate(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	m.AssertExpectations(t)

	expected := []*v4.Frame{
		{
			Metric: v3Frame.Metric,
			Fields: []*v4.Field{
				{

					Name:   "value",
					Values: []float64{1.42},
					Config: &v4.Config{
						Unit: "mm",
						Mappings: []*v4.ValueMapping{
							{From: 1, To: 2, Value: "FOO", Text: "BAR", Color: "yellow"},
						},
					},
					Labels: []*v4.Label{{Key: "foo", Value: "bar"}},
				},
			},
			Timestamps: []*timestamppb.Timestamp{
				ts,
			},
			Meta: &v4.FrameMeta{
				Type:                   v4.FrameMeta_FrameTypeDirectoryListing,
				PreferredVisualization: v4.FrameMeta_VisTypeLogs,
				ExecutedQueryString:    "foo bar baz",
				Notices: []*v4.FrameMeta_Notice{
					{Severity: v4.FrameMeta_Notice_NoticeSeverityWarning, Text: "This is a notice", Link: "https://foo.bar", Inspect: v4.FrameMeta_Notice_InspectTypeError},
				},
			},
		},
	}
	assert.Equal(t, expected, res.Frames)
	assert.Equal(t, "next-please", res.NextToken)
}

func TestAdapter_GetQueryOptions(t *testing.T) {
	req := &v4.GetOptionsRequest{
		QueryType: v4.GetOptionsRequest_GetMetricAggregate,
		SelectedOptions: map[string]string{
			"foo": "bar",
		},
	}
	m := &adapteeMock{}
	v3Response := &v3.GetOptionsResponse{
		Options: []*v3.Option{
			{
				Id:          "foo",
				Description: "bar",
				Type:        v3.Option_Enum,
				EnumValues: []*v3.EnumValue{
					{
						Id:          "1",
						Description: "one",
						Label:       "the first option",
						Default:     true,
					},
				},
			},
		},
	}
	m.On("GetQueryOptions", mock.Anything, &v3.GetOptionsRequest{
		QueryType:       v3.GetOptionsRequest_GetMetricAggregate,
		SelectedOptions: req.SelectedOptions,
	}).Return(v3Response, nil)

	sut := &adapter{
		adaptee: m,
	}

	res, err := sut.GetQueryOptions(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	m.AssertExpectations(t)
	expected := []*v4.Option{
		{
			Id:          "foo",
			Description: "bar",
			Type:        v4.Option_Enum,
			EnumValues: []*v4.EnumValue{
				{
					Id:          "1",
					Description: "one",
					Label:       "the first option",
					Default:     true,
				},
			},
		},
	}
	assert.Equal(t, expected, res.Options)
}
