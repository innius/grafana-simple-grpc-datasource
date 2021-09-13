package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricValue struct {
	*pb.GetMetricValueResponse
}

func (p MetricValue) Frames() (data.Frames, error) {
	f := data.Frames{}

	for _, v := range p.Result {
		timeField := fields.TimeField(1)
		log.DefaultLogger.Debug("MetricValue", "metric", v.Metric.GetId())
		valueField := fields.MetricField("Value", 1)
		//TODO: be a bit more defensive here
		if v.Datapoint != nil {
			timeField.Set(0, getTime(v.Datapoint.GetTimestamp()))
			valueField.Set(0, v.Datapoint.Value.DoubleValue)
		}
		frame := data.NewFrame(v.Metric.GetId(), timeField, valueField)

		f = append(f, frame)
	}
	return f, nil
}
