package framer

import (
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricAggregate struct {
	pb.GetMetricAggregateResponse
	MetricID        string
	AggregationType pb.AggregateType
}

func (p MetricAggregate) Frames() (data.Frames, error) {
	panic("Implement me")
	//	length := len(p.Result)
	//
	//	timeField := fields.TimeField(length)
	//	aggrField := fields.AggregationField(length, aggrTypeAlias(p.AggregationType))
	//	log.DefaultLogger.Debug("MetricAggregate", "value", p.MetricID)
	//	for i, v := range p.Values {
	//		timeField.Set(i, getTime(v.Timestamp))
	//		aggrField.Set(i, v.Value.DoubleValue)
	//	}
	//
	//	frame := data.NewFrame(p.MetricID, timeField, aggrField)
	//
	//	frame.Meta = &data.FrameMeta{
	//		Custom: models.Metadata{
	//			NextToken: p.NextToken,
	//		},
	//	}
	//	return data.Frames{frame}, nil
	//}
	//
	//func aggrTypeAlias(at pb.AggregateType) string {
	//	switch at {
	//	case pb.AggregateType_AVERAGE:
	//		return "avg"
	//	case pb.AggregateType_MIN:
	//		return "min"
	//	case pb.AggregateType_MAX:
	//		return "max"
	//	}
	//	return ""
}
