package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricAggregate struct {
	pb.GetMetricAggregateResponse
	MetricID string
}

func (p MetricAggregate) Frames() (data.Frames, error) {
	length := len(p.Values)

	timeField := fields.TimeField(length)
	valueField := fields.MetricField(p.MetricID, length)
	log.DefaultLogger.Debug("MetricValue", "metric", p.MetricID)
	frame := data.NewFrame(p.MetricID, timeField, valueField)

	for i, v := range p.Values {
		timeField.Set(i, getTime(v.Timestamp))
		valueField.Set(i, v.Value.DoubleValue)
	}

	frame.Meta = &data.FrameMeta{
		Custom: models.Metadata{
			NextToken: p.NextToken,
		},
	}

	return data.Frames{frame}, nil
}
