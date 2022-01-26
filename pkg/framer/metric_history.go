package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricHistory struct {
	*pb.GetMetricHistoryResponse
	Query models.MetricHistoryQuery
}

func (f MetricHistory) Frames() (data.Frames, error) {
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

func (f MetricHistory) FormatDisplayName(metricData *pb.GetMetricHistoryResponse_Data) string {
	return formatDisplayName(FormatDisplayNameInput{
		DisplayName: f.Query.DisplayName,
		MetricID:    metricData.Metric.Id,
		Dimensions:  f.Query.Dimensions,
		Labels:      metricData.GetLabels(),
	})
}

func (f MetricHistory) metricDataToFields(metricData *pb.GetMetricHistoryResponse_Data) []*data.Field {
	length := len(metricData.Series)
	if length == 0 {
		return nil
	}
	metric := metricData.Metric
	timeField := fields.TimeField(length)

	displayName := f.FormatDisplayName(metricData)

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
