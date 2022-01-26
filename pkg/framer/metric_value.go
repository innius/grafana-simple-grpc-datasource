package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricValue struct {
	*pb.GetMetricValueResponse
	Query models.MetricValueQuery
}

func (f MetricValue) Frames() (data.Frames, error) {
	if f.Data == nil {
		return data.Frames{}, nil
	}

	var frames data.Frames

	for i := range f.Data {
		metricData := f.Data[i]
		metric, timestamp := metricData.Metric, metricData.Timestamp

		timeField := fields.TimeField(1)
		timeField.Set(0, getTime(timestamp))

		displayName := f.FormatDisplayName(metricData)

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

func (f MetricValue) FormatDisplayName(metricData *pb.GetMetricValueResponse_Data) string {
	return formatDisplayName(FormatDisplayNameInput{
		DisplayName: f.Query.DisplayName,
		MetricID:    metricData.Metric.Id,
		Dimensions:  f.Query.Dimensions,
		Labels:      metricData.GetLabels(),
	})
}
