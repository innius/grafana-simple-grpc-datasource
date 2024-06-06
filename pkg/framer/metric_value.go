package framer

import (
	"strings"
	"time"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/grafana/grafana-plugin-sdk-go/data"
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

			dataField := convertToSingleDataField(fld)
			dataField.SetConfig(convertToDataFieldConfig(fld.Config, f.FormatDisplayName(metricFrame, fld)))

			fields = append(fields, dataField)
		}

		frame := &data.Frame{
			Name:   metricFrame.Metric,
			Fields: fields,
			Meta:   convertFrameMeta(metricFrame.Meta),
		}
		res = append(res, frame)
	}

	return res, nil
}

func convertToSingleDataField(fld *pb.SingleValueField) *data.Field {
	newField := data.NewField(fld.Name, convertToDataFieldLabels(fld.Labels), convertValue(fld))

	return newField
}

func convertValue(fld *pb.SingleValueField) interface{} {
	if fld.StringValue != "" {
		return []string{fld.StringValue}
	}
	return []float64{fld.Value}
}

func (f MetricValue) FormatDisplayName(frame *pb.GetMetricValueResponse_Frame, fld *pb.SingleValueField) string {
	var args []Arg
	for key, value := range f.Query.Options {
		args = append(args, Arg{Key: strings.ToLower(key), Value: value.Label})
	}
	return formatDisplayName(FormatDisplayNameInput{
		DisplayName: f.Query.DisplayName,
		FieldName:   fld.Name,
		MetricID:    frame.GetMetric(),
		Dimensions:  f.Query.Dimensions,
		Labels:      fld.GetLabels(),
		Args:        args,
	})
}
