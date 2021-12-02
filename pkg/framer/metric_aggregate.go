package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto"
	set "github.com/deckarep/golang-set"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricAggregate struct {
	pb.GetMetricAggregateResponse
	MetricID        string
	AggregationType pb.AggregateType
}

func (p MetricAggregate) Frames() (data.Frames, error) {
	length := len(p.Values)

	log.DefaultLogger.Debug("MetricHistory", "value", p.MetricID)

	set := set.NewSet()

	for _, value := range p.Values[0].Values {
		set.Add(value.Id)
	}

	timeField := fields.TimeField(length)

	result := make(map[string]*data.Field)

	for _, value := range set.ToSlice() {
		newField := fields.AggregationField(len(p.Values), aggrTypeAlias(p.AggregationType)+"_"+value.(string))
		result[value.(string)] = newField
	}

	for index, metricHistoryValue := range p.Values {
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

	frame := &data.Frame{
		Name:   p.MetricID,
		Fields: fields,
	}

	frame.Meta = &data.FrameMeta{
		Custom: models.Metadata{
			NextToken: p.NextToken,
		},
	}
	return data.Frames{frame}, nil
}

func aggrTypeAlias(at pb.AggregateType) string {
	switch at {
	case pb.AggregateType_AVERAGE:
		return "avg"
	case pb.AggregateType_MIN:
		return "min"
	case pb.AggregateType_MAX:
		return "max"
	}
	return ""
}
