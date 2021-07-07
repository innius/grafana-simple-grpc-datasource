package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricValue struct {
	pb.GetMetricValueResponse
	MetricID string
}

func (p MetricValue) Frames() (data.Frames, error) {
	length := 0
	if p.Value != nil {
		length = 1
	}

	timeField := fields.TimeField(length)
	log.DefaultLogger.Debug("MetricValue", "metric", p.MetricID)
	valueField := fields.MetricField("value", length)

	frame := data.NewFrame(p.MetricID, timeField, valueField)

	if p.Value != nil {
		timeField.Set(0, getTime(p.Timestamp))
		//TODO shouldn't we distinguish between nil and 0 ?
		valueField.Set(0, p.Value.DoubleValue)
	}

	return data.Frames{frame}, nil
}
