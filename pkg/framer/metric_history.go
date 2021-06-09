package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricHistory struct {
	pb.GetMetricHistoryResponse
	MetricID string
}

func (p MetricHistory) Frames() (data.Frames, error) {
	length := len(p.Values)

	timeField := fields.TimeField(length)
	valueField := fields.MetricField(p.MetricID, length)
	log.DefaultLogger.Debug("MetricValue", "metric", p.MetricID)
	frame := data.NewFrame(p.MetricID, timeField, valueField)

	for i, v := range p.Values {
		timeField.Set(i, getTime(v.Timestamp))
		//TODO shouldn't we distinguish between nil and 0 ?
		valueField.Set(i, v.Value.DoubleValue)
	}

	return data.Frames{frame}, nil
}
