package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricValue struct {
	pb.GetMetricValueResponse
	models.MetricValueQuery
}

func (p MetricValue) Frames() (data.Frames, error) {
	length := 0
	if p.Value != nil {
		length = 1
	}

	timeField := fields.TimeField(length)
	log.DefaultLogger.Debug("MetricValue", "metric", p.MetricId)
	valueField := fields.MetricField("value", length)
	if p.DisplayName != "" {
		valueField.Config = &data.FieldConfig{
			DisplayNameFromDS: p.FormatDisplayName(),
		}
	}
	frame := data.NewFrame(p.MetricId, timeField, valueField)

	if p.Value != nil {
		timeField.Set(0, getTime(p.Timestamp))
		valueField.Set(0, p.Value.DoubleValue)
	}

	return data.Frames{frame}, nil
}
