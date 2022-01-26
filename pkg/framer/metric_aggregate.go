package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricAggregate struct {
	*pb.GetMetricAggregateResponse
	Query models.MetricBaseQuery
	pb.AggregateType
}

func (f MetricAggregate) AggregateTypeAlias() string {
	switch f.AggregateType {
	case pb.AggregateType_AVERAGE:
		return "avg"
	case pb.AggregateType_MIN:
		return "min"
	case pb.AggregateType_MAX:
		return "max"
	case pb.AggregateType_COUNT:
		return "count"
	}
	return ""
}

func (f MetricAggregate) Frames() (data.Frames, error) {
	if f.Data == nil {
		return data.Frames{}, nil
	}

	var frames data.Frames

	for i := range f.Data {
		metricData := f.Data[i]
		frame := &data.Frame{
			Name:   metricData.Metric.Id,
			Fields: f.metricDataToFields(metricData),
		}
		frames = append(frames, frame)
	}

	// add metadata -> add next token to the first frame (this is how other datasource plugins are doing this)
	if len(frames) > 0 {
		frame := frames[0]
		frame.Meta = &data.FrameMeta{
			Custom: models.Metadata{
				NextToken: f.NextToken,
			},
		}
	}

	return frames, nil
}

func (f MetricAggregate) FormatDisplayName(metricData *pb.GetMetricAggregateResponse_Data) string {
	return formatDisplayName(FormatDisplayNameInput{
		DisplayName: f.Query.DisplayName,
		MetricID:    metricData.Metric.Id,
		Dimensions:  f.Query.Dimensions,
		Labels:      metricData.GetLabels(),
		Args: []Arg{{
			Key:   "aggregate",
			Value: f.AggregateTypeAlias(),
		}},
	})
}

func (f *MetricAggregate) metricDataToFields(metricData *pb.GetMetricAggregateResponse_Data) []*data.Field {
	length := len(metricData.Series)
	if length == 0 {
		return nil
	}
	metric := metricData.Metric
	timeField := fields.TimeField(length)

	displayName := f.FormatDisplayName(metricData)

	dataField := newDataFieldForMetric2(metric, f.AggregateTypeAlias(), metricData.Labels, displayName, length)

	for index, v := range metricData.Series {
		timeField.Set(index, getTime(v.Timestamp))
		var value float64
		if v.Value != nil {
			value = v.Value.DoubleValue
		}
		dataField.Set(index, value)
	}
	return []*data.Field{
		timeField,
		dataField,
	}
}
