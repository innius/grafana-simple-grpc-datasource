package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricValue struct {
	*pb.GetMetricValueResponse
	models.MetricValueQuery
}

func (p MetricValue) Frames() (data.Frames, error) {
	if p.Data == nil {
		return data.Frames{}, nil
	}

	var frames data.Frames

	for i := range p.Data {
		metricData := p.Data[i]
		metric, timestamp := metricData.Metric, metricData.Timestamp

		timeField := fields.TimeField(1)
		timeField.Set(0, getTime(timestamp))

		var displayName *string
		if p.DisplayName != "" {
			s := p.FormatDisplayName(metric.Id, metricData.Labels)
			displayName = &s
		}

		dataField := newDataFieldForMetric(metric, metricData.Labels, displayName, 1)
		var value float64
		if metricData.Value != nil {
			value = metricData.Value.DoubleValue
		}
		dataField.Set(0, value)

		frame := &data.Frame{
			Name: metric.Id,
			Fields: []*data.Field{
				timeField, dataField,
			},
		}
		frames = append(frames, frame)
	}

	return frames, nil
}
