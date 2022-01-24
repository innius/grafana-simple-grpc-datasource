package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"fmt"
	set "github.com/deckarep/golang-set"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricAggregate struct {
	models.MetricAggregateQuery
	pb.GetMetricAggregateResponse
	AggregationType pb.AggregateType
}

func (p MetricAggregate) Frames() (data.Frames, error) {
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

func (p MetricAggregate) seriesToFields(metricID string, series []*pb.GetMetricAggregateResponse_Data_TimeSeries) []*data.Field {
	set := set.NewSet()

	backend.Logger.Info(fmt.Sprintf("The series: %+v", series))
	if len(series) == 0 {
		return nil
	}
	for _, s := range series[0].Values {
		set.Add(aggrTypeAlias(s.AggregateType))
	}
	backend.Logger.Info(fmt.Sprintf("The set: %+v", set))

	timeField := fields.TimeField(len(series))

	result := make(map[string]*data.Field)

	for _, value := range set.ToSlice() {
		key := value.(string)
		newField := fields.MetricField(key, len(series))
		if p.DisplayName != "" {
			newField.Config = &data.FieldConfig{
				DisplayNameFromDS: p.FormatDisplayName(metricID, value.(string)),
			}
		}
		result[key] = newField
	}

	backend.Logger.Info(fmt.Sprintf("The dict: %+v", result))

	for index, metricHistoryValue := range series {
		timeField.Set(index, getTime(metricHistoryValue.Timestamp))
		for _, value := range metricHistoryValue.Values {
			backend.Logger.Info(fmt.Sprintf("The key: %s", aggrTypeAlias(value.AggregateType)))

			result[aggrTypeAlias(value.AggregateType)].Set(index, value.DoubleValue)
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

func aggrTypeAlias(at pb.AggregateType) string {
	switch at {
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
