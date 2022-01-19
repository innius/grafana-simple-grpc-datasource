package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricAggregate struct {
	models.MetricAggregateQuery
	pb.GetMetricAggregateResponse
	AggregationType pb.AggregateType
}

func (p MetricAggregate) Frames() (data.Frames, error) {
	length := len(p.Values)
	timeField := fields.TimeField(length)
	valueField := fields.AggregationField(length, aggrTypeAlias(p.AggregationType))

	if p.DisplayName != "" {
		valueField.Config = &data.FieldConfig{
			DisplayNameFromDS: p.FormatDisplayName(),
		}
	}

	log.DefaultLogger.Debug("MetricAggregate", "value", p.MetricId)
	for i, v := range p.Values {
		timeField.Set(i, getTime(v.Timestamp))
		valueField.Set(i, v.Value.DoubleValue)
	}

	frame := data.NewFrame(p.MetricId, timeField, valueField)

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
