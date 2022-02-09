package framer

import (
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
	return convertToDataFrames(f), nil
}

func (f MetricAggregate) FormatDisplayName(frame *pb.Frame, fld *pb.Field) string {
	return formatDisplayName(FormatDisplayNameInput{
		DisplayName: f.Query.DisplayName,
		FieldName:   fld.Name,
		MetricID:    frame.Metric,
		Dimensions:  f.Query.Dimensions,
		Labels:      fld.GetLabels(),
		Args: []Arg{{
			Key:   "aggregate",
			Value: f.AggregateTypeAlias(),
		}},
	})
}
