package v2

import (
	"context"
	"strconv"

	v2 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/samber/lo"
	"google.golang.org/grpc"
)

type adapter struct {
	v2Client v2.GrafanaQueryAPIClient
}

// Returns a list of all available dimensions
func (adapter *adapter) ListDimensionKeys(ctx context.Context, in *v3.ListDimensionKeysRequest, opts ...grpc.CallOption) (*v3.ListDimensionKeysResponse, error) {

	inv1 := &v2.ListDimensionKeysRequest{
		Filter: in.Filter,
	}
	res, err := adapter.v2Client.ListDimensionKeys(ctx, inv1, opts...)
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

// Returns a list of all dimension values for a certain dimension
func (adapter *adapter) ListDimensionValues(ctx context.Context, in *v3.ListDimensionValuesRequest, opts ...grpc.CallOption) (*v3.ListDimensionValuesResponse, error) {

	inv1 := &v2.ListDimensionValuesRequest{
		DimensionKey: in.DimensionKey,
		Filter:       in.Filter,
	}
	res, err := adapter.v2Client.ListDimensionValues(ctx, inv1, opts...)
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

func toV2Dimensions(in []*v3.Dimension) []*v2.Dimension {
	res := make([]*v2.Dimension, len(in))
	for i, v := range in {
		res[i] = &v2.Dimension{
			Key:   v.Key,
			Value: v.Value,
		}
	}
	return res
}

// Returns all metrics from the system
func (adapter *adapter) ListMetrics(ctx context.Context, in *v3.ListMetricsRequest, opts ...grpc.CallOption) (*v3.ListMetricsResponse, error) {
	inv1 := &v2.ListMetricsRequest{
		Dimensions: toV2Dimensions(in.Dimensions),
		Filter:     in.Filter,
	}
	res, err := adapter.v2Client.ListMetrics(ctx, inv1, opts...)
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

// Gets the options for the specified query type
func (adapter *adapter) GetQueryOptions(ctx context.Context, in *v3.GetOptionsRequest, opts ...grpc.CallOption) (*v3.GetOptionsResponse, error) {
	if in.QueryType == v3.GetOptionsRequest_GetMetricAggregate {
		return &v3.GetOptionsResponse{
			Options: []*v3.Option{
				{
					Id:          "Aggregate",
					Label:       "Aggregate",
					Type:        v3.Option_Enum,
					Description: "Selects the aggregate for metric values",
					EnumValues: []*v3.EnumValue{
						{Label: "Average", Description: "Average value aggregate", Id: strconv.Itoa(int(v2.AggregateType_AVERAGE))},
						{Label: "Min", Description: "Min value aggregate", Id: strconv.Itoa(int(v2.AggregateType_MIN))},
						{Label: "Max", Description: "Max value aggregate", Id: strconv.Itoa(int(v2.AggregateType_MAX))},
						{Label: "Count", Description: "Count value aggregate", Id: strconv.Itoa(int(v2.AggregateType_COUNT))},
					}},
			},
		}, nil
	}
	return &v3.GetOptionsResponse{}, nil
}

// Gets the last known value for one or more metrics
func (adapter *adapter) GetMetricValue(ctx context.Context, in *v3.GetMetricValueRequest, opts ...grpc.CallOption) (*v3.GetMetricValueResponse, error) {
	if len(in.Metrics) == 0 {
		return &v3.GetMetricValueResponse{}, nil
	}
	inv1 := &v2.GetMetricValueRequest{
		Dimensions: toV2Dimensions(in.Dimensions),
		Metrics:    in.Metrics,
	}
	res, err := adapter.v2Client.GetMetricValue(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}

	var frames []*v3.GetMetricValueResponse_Frame = make([]*v3.GetMetricValueResponse_Frame, len(res.Frames))

	frames = lo.Map(res.Frames, func(frame *v2.GetMetricValueResponse_Frame, _ int) *v3.GetMetricValueResponse_Frame {
		return &v3.GetMetricValueResponse_Frame{
			Metric:    frame.Metric,
			Timestamp: frame.Timestamp,
			Fields: lo.Map(frame.Fields, func(f *v2.SingleValueField, _ int) *v3.SingleValueField {
				return &v3.SingleValueField{
					Name: f.Name,
					Labels: lo.Map(f.Labels, func(l *v2.Label, _ int) *v3.Label {
						return &v3.Label{Key: l.Key, Value: l.Value}
					}),
					Value:  f.Value,
					Config: toV3Config(f.GetConfig()),
				}
			}),
			Meta: toV3Meta(frame.GetMeta()),
		}
	})
	return &v3.GetMetricValueResponse{
		Frames: frames,
	}, nil
}

// Gets the history for one or more metrics
func (adapter *adapter) GetMetricHistory(ctx context.Context, in *v3.GetMetricHistoryRequest, opts ...grpc.CallOption) (*v3.GetMetricHistoryResponse, error) {
	inv1 := &v2.GetMetricHistoryRequest{
		Dimensions:    toV2Dimensions(in.Dimensions),
		Metrics:       in.Metrics,
		StartDate:     in.StartDate,
		EndDate:       in.EndDate,
		MaxItems:      in.MaxItems,
		TimeOrdering:  v2.TimeOrdering(in.TimeOrdering),
		StartingToken: in.StartingToken,
	}
	res, err := adapter.v2Client.GetMetricHistory(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	frames := lo.Map(res.Frames, toV3Frame)
	return &v3.GetMetricHistoryResponse{
		Frames:    frames,
		NextToken: res.NextToken,
	}, nil
}

const aggregateTypeOptionKey = "aggregateType"

// Gets the history for one or more metrics
func (adapter *adapter) GetMetricAggregate(ctx context.Context, in *v3.GetMetricAggregateRequest, opts ...grpc.CallOption) (*v3.GetMetricAggregateResponse, error) {
	var aggregateType v2.AggregateType
	switch in.GetOptions()[aggregateTypeOptionKey] {
	case "0":
		aggregateType = v2.AggregateType_AVERAGE
	case "1":
		aggregateType = v2.AggregateType_MIN
	case "2":
		aggregateType = v2.AggregateType_MAX
	case "3":
		aggregateType = v2.AggregateType_COUNT
	default:
		aggregateType = v2.AggregateType_AVERAGE
	}

	inv1 := &v2.GetMetricAggregateRequest{
		Dimensions:    toV2Dimensions(in.Dimensions),
		Metrics:       in.Metrics,
		StartDate:     in.StartDate,
		EndDate:       in.EndDate,
		MaxItems:      in.MaxItems,
		TimeOrdering:  v2.TimeOrdering(in.TimeOrdering),
		StartingToken: in.StartingToken,
		AggregateType: aggregateType,
		IntervalMs:    in.IntervalMs,
	}
	res, err := adapter.v2Client.GetMetricAggregate(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	frames := lo.Map(res.Frames, toV3Frame)
	return &v3.GetMetricAggregateResponse{
		Frames:    frames,
		NextToken: res.NextToken,
	}, nil
}

func toV3Frame(frame *v2.Frame, _ int) *v3.Frame {
	return &v3.Frame{
		Metric:     frame.Metric,
		Timestamps: frame.Timestamps,
		Fields: lo.Map(frame.Fields, func(f *v2.Field, _ int) *v3.Field {
			return &v3.Field{
				Labels: lo.Map(f.Labels, func(l *v2.Label, _ int) *v3.Label {
					return &v3.Label{Key: l.Key, Value: l.Value}
				}),
				Config: toV3Config(f.GetConfig()),
				Name:   f.Name,
				Values: f.Values,
			}
		}),
		Meta: toV3Meta(frame.Meta),
	}
}

func toV3Meta(m *v2.FrameMeta) *v3.FrameMeta {
	if m == nil {
		return nil
	}
	return &v3.FrameMeta{
		Type:                   v3.FrameMeta_FrameType(m.Type),
		PreferredVisualization: v3.FrameMeta_VisType(m.PreferredVisualization),
		ExecutedQueryString:    m.ExecutedQueryString,
		Notices: lo.Map(m.Notices, func(notice *v2.FrameMeta_Notice, _ int) *v3.FrameMeta_Notice {
			return &v3.FrameMeta_Notice{
				Severity: v3.FrameMeta_Notice_NoticeSeverity(notice.Severity),
				Text:     notice.Text,
				Link:     notice.Link,
				Inspect:  v3.FrameMeta_Notice_InspectType(notice.Inspect),
			}
		}),
	}
}

func toV3Config(cfg *v2.Config) *v3.Config {
	if cfg == nil {
		return nil
	}
	return &v3.Config{
		Unit: cfg.Unit,
		Mappings: lo.Map(cfg.Mappings, func(m *v2.ValueMapping, _ int) *v3.ValueMapping {
			return &v3.ValueMapping{
				From:  m.From,
				To:    m.To,
				Value: m.Value,
				Text:  m.Text,
				Color: m.Color,
			}
		}),
	}
}
