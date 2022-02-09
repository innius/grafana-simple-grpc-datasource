package framer

import (
	fields2 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func convertToDataField(fld *pb.Field) *data.Field {
	newField := data.NewField(fld.Name, convertToDataFieldLabels(fld.Labels), convertValues(fld))

	return newField
}

func convertToDataFieldConfig(config *pb.Config, formatDisplayName string) *data.FieldConfig {
	var cfg *data.FieldConfig
	if config != nil {
		cfg = &data.FieldConfig{
			Unit: config.Unit,
		}
	}
	if formatDisplayName != "" {
		if cfg == nil {
			cfg = &data.FieldConfig{}
		}
		cfg.DisplayNameFromDS = formatDisplayName
	}

	return cfg
}

func convertToDataFieldLabels(labels []*pb.Label) data.Labels {
	var dataLabels = make(data.Labels, len(labels))

	for i := range labels {
		label := labels[i]
		dataLabels[label.Key] = label.Value
	}
	return dataLabels
}

func convertValues(fld *pb.Field) interface{} {
	return fld.Values
}

type framesResponse interface {
	GetFrames() []*pb.Frame
	GetNextToken() string
	FormatDisplayName(frame *pb.Frame, fld *pb.Field) string
}

func convertToDataFrames(response framesResponse) data.Frames {
	if response == nil {
		return data.Frames{}
	}

	var res data.Frames
	frames := response.GetFrames()

	for i := range frames {
		metricFrame := frames[i]

		timeField := fields2.TimeField(len(metricFrame.Timestamps))
		for idx := range metricFrame.Timestamps {
			timeField.Set(idx, metricFrame.Timestamps[idx].AsTime())
		}

		fields := data.Fields{
			timeField,
		}

		for idx := range metricFrame.Fields {
			fld := metricFrame.Fields[idx]
			dataField := convertToDataField(fld)
			dataField.SetConfig(convertToDataFieldConfig(fld.Config, response.FormatDisplayName(metricFrame, fld)))

			fields = append(fields, dataField)
		}

		frame := &data.Frame{
			Name:   metricFrame.Metric,
			Fields: fields,
		}
		res = append(res, frame)
	}

	// add metadata -> add next token to the first frame (this is how other datasource plugins are doing this)
	if len(res) > 0 {
		frame := res[0]
		frame.Meta = &data.FrameMeta{
			Custom: models.Metadata{
				NextToken: response.GetNextToken(),
			},
		}
	}

	return res
}
