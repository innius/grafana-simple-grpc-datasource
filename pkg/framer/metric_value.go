package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricValue struct {
	pb.GetMetricValueResponse
	models.MetricValueQuery
}

func (p MetricValue) Frames() (data.Frames, error) {
	length := 0
	if p.Values != nil {
		length = 1
	}
	log.DefaultLogger.Debug("MetricValue", "metric", p.MetricId)

	timeField := fields.TimeField(length)
	timeField.Set(0, getTime(p.Timestamp))

	result := []*data.Field{
		timeField,
	}

	if len(p.Values) == 1 {
		metricValue := p.Values[0]
		valueField := fields.MetricField("value", length)
		if p.DisplayName != "" {
			valueField.Config = &data.FieldConfig{
				DisplayNameFromDS: p.FormatDisplayName(),
			}
		}
		valueField.Set(0, metricValue.DoubleValue)
		result = append(result, valueField)
	} else {
		for _, metricValue := range p.Values {
			newField := fields.MetricField(metricValue.Id, 1)
			newField.Set(0, metricValue.DoubleValue)
			result = append(result, newField)
		}
	}

	frame := &data.Frame{
		Name:   p.MetricId,
		Fields: result,
	}

	return data.Frames{frame}, nil
}
