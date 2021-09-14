package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricHistory struct {
	*pb.GetMetricHistoryResponse
}

func (p MetricHistory) Frames() (data.Frames, error) {
	frames := data.Frames{}

	for _, v := range p.Result {
		length := len(v.Values)

		timeField := fields.TimeField(length)
		valueField := fields.MetricField("Value", length)

		log.DefaultLogger.Debug("MetricValue", "metric", v.Metric.GetId())

		datapoints := v.GetValues()
		log.DefaultLogger.Debug(fmt.Sprintf("Datapoints; %d", len(datapoints)))
		for i := range datapoints {
			dp := datapoints[i]
			timeField.Set(i, getTime(dp.GetTimestamp()))
			if dp.Value != nil {
				if dp.Value != nil {
					valueField.Set(i, dp.Value.DoubleValue)
				}
			}
		}

		frame := data.NewFrame(v.Metric.GetId(), timeField, valueField)
		frames = append(frames, frame)
	}

	meta := &data.FrameMeta{
		Custom: models.Metadata{
			NextToken: p.NextToken,
		},
	}
	// Needs a frame for the metadata... even if just error
	if len(frames) < 1 {
		frames = append(frames, data.NewFrame(""))
	}
	frame := frames[0]
	if frame.Meta == nil {
		frame.Meta = &data.FrameMeta{}
	}
	frame.Meta = meta
	return frames, nil
}
