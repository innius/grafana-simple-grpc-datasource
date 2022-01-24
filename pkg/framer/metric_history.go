package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	set "github.com/deckarep/golang-set"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricHistory struct {
	pb.GetMetricHistoryResponse
	models.MetricHistoryQuery
}

func (p MetricHistory) seriesToFields(metricID string, series []*pb.GetMetricHistoryResponse_Data_TimeSeries) []*data.Field {
	set := set.NewSet()

	for _, s := range series[0].Values {
		set.Add(s.Id)
	}

	timeField := fields.TimeField(len(series))

	result := make(map[string]*data.Field)

	for _, value := range set.ToSlice() {
		newField := fields.MetricField(value.(string), len(series))
		if p.DisplayName != "" {
			newField.Config = &data.FieldConfig{
				DisplayNameFromDS: p.FormatDisplayName(metricID, value.(string)),
			}
		}
		result[value.(string)] = newField
	}

	for index, metricHistoryValue := range series {
		timeField.Set(index, getTime(metricHistoryValue.Timestamp))
		for _, value := range metricHistoryValue.Values {
			result[value.Id].Set(index, value.DoubleValue)
		}
	}
	fields := []*data.Field{
		timeField,
	}

	for _, field := range result {
		fields = append(fields, field)
	}

	return fields
}

func (p MetricHistory) Frames() (data.Frames, error) {
	if p.Data == nil {
		return data.Frames{}, nil
	}

	var frames data.Frames

	for i := range p.Data {
		res := p.Data[i]
		metric, series := res.Metric, res.Series
		frame := &data.Frame{
			Name:   metric.Id,
			Fields: p.seriesToFields(metric.Id, series),
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
