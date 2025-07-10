package v3

import (
	"context"

	"github.com/samber/lo"
	"google.golang.org/grpc"

	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	v4 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v4"
)

type adapter struct {
	adaptee v3.GrafanaQueryAPIClient
}

// Returns a list of all available dimensions
func (adapter *adapter) ListDimensionKeys(ctx context.Context, in *v4.ListDimensionKeysRequest, opts ...grpc.CallOption) (*v4.ListDimensionKeysResponse, error) {

	inv1 := &v3.ListDimensionKeysRequest{
		Filter: in.Filter,
	}
	res, err := adapter.adaptee.ListDimensionKeys(ctx, inv1, opts...)
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

// Returns a list of all dimension values for a certain dimension
func (adapter *adapter) ListDimensionValues(ctx context.Context, in *v4.ListDimensionValuesRequest, opts ...grpc.CallOption) (*v4.ListDimensionValuesResponse, error) {

	inv1 := &v3.ListDimensionValuesRequest{
		DimensionKey: in.DimensionKey,
		Filter:       in.Filter,
	}
	res, err := adapter.adaptee.ListDimensionValues(ctx, inv1, opts...)
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

func convertDimensions(in []*v4.Dimension) []*v3.Dimension {
	res := make([]*v3.Dimension, len(in))
	for i, v := range in {
		res[i] = &v3.Dimension{
			Key:   v.Key,
			Value: v.Value,
		}
	}
	return res
}

// Returns all metrics from the system
func (adapter *adapter) ListMetrics(ctx context.Context, in *v4.ListMetricsRequest, opts ...grpc.CallOption) (*v4.ListMetricsResponse, error) {
	inv1 := &v3.ListMetricsRequest{
		Dimensions: convertDimensions(in.Dimensions),
		Filter:     in.Filter,
	}
	res, err := adapter.adaptee.ListMetrics(ctx, inv1, opts...)
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

// Gets the options for the specified query type
func (adapter *adapter) GetQueryOptions(ctx context.Context, in *v4.GetOptionsRequest, opts ...grpc.CallOption) (*v4.GetOptionsResponse, error) {
	res, err := adapter.adaptee.GetQueryOptions(ctx, &v3.GetOptionsRequest{
		QueryType:       v3.GetOptionsRequest_QueryType(in.QueryType),
		SelectedOptions: in.SelectedOptions,
	}, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return &v4.GetOptionsResponse{}, nil
	}
	options := lo.Map(res.Options, func(o *v3.Option, _ int) *v4.Option {
		return &v4.Option{
			Id:          o.Id,
			Description: o.Description,
			Type:        v4.Option_Type(o.Type),
			EnumValues: lo.Map(o.EnumValues, func(ev *v3.EnumValue, _ int) *v4.EnumValue {
				return &v4.EnumValue{
					Id:          ev.Id,
					Description: ev.Description,
					Label:       ev.Label,
					Default:     ev.Default,
				}
			}),
			Required: o.Required,
			Label:    o.Label,
		}
	})
	return &v4.GetOptionsResponse{Options: options}, nil
}

// Gets the last known value for one or more metrics
func (adapter *adapter) GetMetricValue(ctx context.Context, in *v4.GetMetricValueRequest, opts ...grpc.CallOption) (*v4.GetMetricValueResponse, error) {
	if len(in.Metrics) == 0 {
		return &v4.GetMetricValueResponse{}, nil
	}
	inv1 := &v3.GetMetricValueRequest{
		Dimensions: convertDimensions(in.Dimensions),
		Metrics:    in.Metrics,
		Options:    in.Options,
		StartDate:  in.StartDate,
		EndDate:    in.EndDate,
	}
	res, err := adapter.adaptee.GetMetricValue(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}

	frames := lo.Map(res.Frames, func(frame *v3.GetMetricValueResponse_Frame, _ int) *v4.GetMetricValueResponse_Frame {
		return &v4.GetMetricValueResponse_Frame{
			Metric:    frame.Metric,
			Timestamp: frame.Timestamp,
			Fields: lo.Map(frame.Fields, func(f *v3.SingleValueField, _ int) *v4.SingleValueField {
				return &v4.SingleValueField{
					Name: f.Name,
					Labels: lo.Map(f.Labels, func(l *v3.Label, _ int) *v4.Label {
						return &v4.Label{Key: l.Key, Value: l.Value}
					}),
					Value:  f.Value,
					Config: convertConfig(f.GetConfig()),
				}
			}),
			Meta: convertMeta(frame.GetMeta()),
		}
	})
	return &v4.GetMetricValueResponse{
		Frames: frames,
	}, nil
}

// Gets the history for one or more metrics
func (adapter *adapter) GetMetricHistory(ctx context.Context, in *v4.GetMetricHistoryRequest, opts ...grpc.CallOption) (*v4.GetMetricHistoryResponse, error) {
	inv1 := &v3.GetMetricHistoryRequest{
		Dimensions:    convertDimensions(in.Dimensions),
		Options:       in.Options,
		Metrics:       in.Metrics,
		StartDate:     in.StartDate,
		EndDate:       in.EndDate,
		MaxItems:      in.MaxItems,
		TimeOrdering:  v3.TimeOrdering(in.TimeOrdering),
		StartingToken: in.StartingToken,
	}
	res, err := adapter.adaptee.GetMetricHistory(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	frames := lo.Map(res.Frames, convertFrame)
	return &v4.GetMetricHistoryResponse{
		Frames:    frames,
		NextToken: res.NextToken,
	}, nil
}

// Gets the history for one or more metrics
func (adapter *adapter) GetMetricAggregate(ctx context.Context, in *v4.GetMetricAggregateRequest, opts ...grpc.CallOption) (*v4.GetMetricAggregateResponse, error) {
	inv1 := &v3.GetMetricAggregateRequest{
		Dimensions:    convertDimensions(in.Dimensions),
		Metrics:       in.Metrics,
		StartDate:     in.StartDate,
		EndDate:       in.EndDate,
		MaxItems:      in.MaxItems,
		TimeOrdering:  v3.TimeOrdering(in.TimeOrdering),
		StartingToken: in.StartingToken,
		IntervalMs:    in.IntervalMs,
		Options:       in.Options,
	}
	res, err := adapter.adaptee.GetMetricAggregate(ctx, inv1, opts...)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	frames := lo.Map(res.Frames, convertFrame)
	return &v4.GetMetricAggregateResponse{
		Frames:    frames,
		NextToken: res.NextToken,
	}, nil
}

func convertFrame(frame *v3.Frame, _ int) *v4.Frame {
	return &v4.Frame{
		Metric:     frame.Metric,
		Timestamps: frame.Timestamps,
		Fields: lo.Map(frame.Fields, func(f *v3.Field, _ int) *v4.Field {
			return &v4.Field{
				Labels: lo.Map(f.Labels, func(l *v3.Label, _ int) *v4.Label {
					return &v4.Label{Key: l.Key, Value: l.Value}
				}),
				Config: convertConfig(f.GetConfig()),
				Name:   f.Name,
				Values: f.Values,
			}
		}),
		Meta: convertMeta(frame.Meta),
	}
}

func convertMeta(m *v3.FrameMeta) *v4.FrameMeta {
	if m == nil {
		return nil
	}
	return &v4.FrameMeta{
		Type:                   v4.FrameMeta_FrameType(m.Type),
		PreferredVisualization: v4.FrameMeta_VisType(m.PreferredVisualization),
		ExecutedQueryString:    m.ExecutedQueryString,
		Notices: lo.Map(m.Notices, func(notice *v3.FrameMeta_Notice, _ int) *v4.FrameMeta_Notice {
			return &v4.FrameMeta_Notice{
				Severity: v4.FrameMeta_Notice_NoticeSeverity(notice.Severity),
				Text:     notice.Text,
				Link:     notice.Link,
				Inspect:  v4.FrameMeta_Notice_InspectType(notice.Inspect),
			}
		}),
	}
}

func convertConfig(cfg *v3.Config) *v4.Config {
	if cfg == nil {
		return nil
	}
	return &v4.Config{
		Unit: cfg.Unit,
		Mappings: lo.Map(cfg.Mappings, func(m *v3.ValueMapping, _ int) *v4.ValueMapping {
			return &v4.ValueMapping{
				From:  m.From,
				To:    m.To,
				Value: m.Value,
				Text:  m.Text,
				Color: m.Color,
			}
		}),
	}
}

func (adapter *adapter) GetStreamingQueryConfiguration(ctx context.Context, in *v4.GetStreamingQueryConfigurationRequest, opts ...grpc.CallOption) (*v4.GetStreamingQueryConfigurationResponse, error) {
	return nil, nil
}
