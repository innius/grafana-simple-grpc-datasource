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
	if p.Values != nil {
		length = 1
	}

	timeField := fields.TimeField(length)
	timeField.Set(0, getTime(p.Timestamp))

	log.DefaultLogger.Debug("MetricValue", "metric", p.MetricID)

	result := []*data.Field{
		timeField,
	}

	for _, metricValue := range p.Values {
		newField := fields.MetricField(metricValue.Id, 1)
		newField.Set(0, metricValue.DoubleValue)
		result = append(result, newField)
	}

	frame := &data.Frame{
		Name:   p.MetricID,
		Fields: result,
	}

	return data.Frames{frame}, nil
}
