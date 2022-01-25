package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricHistory struct {
	*pb.GetMetricHistoryResponse
	models.MetricHistoryQuery
}

func (p MetricHistory) Frames() (data.Frames, error) {
	if p.Data == nil {
		return data.Frames{}, nil
	}

	var frames data.Frames

	for i := range p.Data {
		metricData := p.Data[i]
		frame := &data.Frame{
			Name:   metricData.Metric.Id,
			Fields: p.metricDataToFields(metricData),
		}
		frames = append(frames, frame)
	}

	// add metadata -> add next token to the first frame (this is how other datasource plugins are doing this)
	if len(frames) > 0 {
		frame := frames[0]
		frame.Meta = &data.FrameMeta{
			Custom: models.Metadata{
				NextToken: p.NextToken,
			},
		}
	}

	return frames, nil
}

func newDataFieldForMetric(metric *pb.Metric, labels []*pb.Label, displayName *string, length int) *data.Field {
	dataField := fields.MetricField(metric.Id, length)
	dataField.Config = &data.FieldConfig{}
	if displayName != nil {
		dataField.Config.DisplayNameFromDS = *displayName
	}
	if metric.Unit != "" {
		dataField.Config.Unit = metric.Unit
	}

	if len(labels) > 0 {
		dataField.Labels = data.Labels{}
		for i := range labels {
			label := labels[i]
			dataField.Labels[label.Key] = label.Value
		}
	}
	return dataField
}

func (p MetricHistory) metricDataToFields(metricData *pb.GetMetricHistoryResponse_Data) []*data.Field {
	length := len(metricData.Series)
	if length == 0 {
		return nil
	}
	metric := metricData.Metric
	timeField := fields.TimeField(length)

	var displayName *string
	if p.DisplayName != "" {
		s := p.FormatDisplayName(metric.Id, metricData.Labels)
		displayName = &s
	}

	dataField := newDataFieldForMetric(metric, metricData.Labels, displayName, length)

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
