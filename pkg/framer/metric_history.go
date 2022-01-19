package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricHistory struct {
	pb.GetMetricHistoryResponse
	models.MetricHistoryQuery
}

func (p MetricHistory) Frames() (data.Frames, error) {
	length := len(p.Values)

	timeField := fields.TimeField(length)
	valueField := fields.MetricField("Value", length)
	if p.DisplayName != "" {
		valueField.Config = &data.FieldConfig{
			DisplayNameFromDS: p.FormatDisplayName(),
		}
	}
	log.DefaultLogger.Debug("MetricHistory", "value", p.MetricId)

	frame := data.NewFrame(p.MetricId, timeField, valueField)

	frame.Meta = &data.FrameMeta{
		Custom: models.Metadata{
			NextToken: p.NextToken,
		},
	}
	for i, v := range p.Values {
		timeField.Set(i, getTime(v.Timestamp))
		valueField.Set(i, v.Value.DoubleValue)
	}

	return data.Frames{frame}, nil
}
