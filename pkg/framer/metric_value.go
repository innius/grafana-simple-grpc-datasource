package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricValue struct {
	pb.GetMetricValueResponse
	models.MetricValueQuery
}

func (p MetricValue) Frames() (data.Frames, error) {
	if p.Data == nil {
		return data.Frames{}, nil
	}

	var frames data.Frames

	for i := range p.Data {
		res := p.Data[i]
		metric, timestamp, values := res.Metric, res.Timestamp, res.Values

		log.DefaultLogger.Info(fmt.Sprintf("the data %+v", res))
		timeField := fields.TimeField(1)
		timeField.Set(0, getTime(timestamp))

		result := []*data.Field{
			timeField,
		}

		for _, metricValue := range values {
			newField := fields.MetricField(metricValue.Id, 1)
			if p.DisplayName != "" {
				newField.Config = &data.FieldConfig{
					DisplayNameFromDS: p.FormatDisplayName(metric.Id, metricValue.Id),
				}
			}
			newField.Set(0, metricValue.DoubleValue)
			result = append(result, newField)
		}

		frame := &data.Frame{
			Name:   metric.Id,
			Fields: result,
		}
		frames = append(frames, frame)
	}

	log.DefaultLogger.Debug("MetricValue", "metric", p.Metrics)

	return frames, nil
}
