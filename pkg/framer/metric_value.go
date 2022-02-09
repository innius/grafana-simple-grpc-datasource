package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"time"
)

type MetricValue struct {
	*pb.GetMetricValueResponse
	Query models.MetricValueQuery
}

func (f MetricValue) Frames() (data.Frames, error) {
	frames := f.GetFrames()
	if frames == nil {
		return data.Frames{}, nil
	}

	var res data.Frames

	for i := range frames {
		metricFrame := frames[i]

		fields := data.Fields{
			data.NewField("time", nil, []time.Time{metricFrame.Timestamp.AsTime()}),
		}

		for idx := range metricFrame.Fields {
			fld := metricFrame.Fields[idx]

			dataField := data.NewField(fld.Name, convertToDataFieldLabels(fld.Labels), []float64{fld.Value})
			dataField.SetConfig(convertToDataFieldConfig(fld.Config, f.FormatDisplayName(metricFrame, fld)))

			fields = append(fields, dataField)
		}

		frame := &data.Frame{
			Name:   metricFrame.Metric,
			Fields: fields,
		}
		res = append(res, frame)
	}

	return res, nil
}

func (f MetricValue) FormatDisplayName(frame *pb.GetMetricValueResponse_Frame, fld *pb.SingleValueField) string {
	return formatDisplayName(FormatDisplayNameInput{
		DisplayName: f.Query.DisplayName,
		FieldName:   fld.Name,
		MetricID:    frame.GetMetric(),
		Dimensions:  f.Query.Dimensions,
		Labels:      fld.GetLabels(),
	})
}
