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

	mappings := []data.ValueMapping{}
	for _, v := range config.GetMappings() {
		switch {
		case v.GetValue() != "":
			m := data.ValueMapper{}
			m[v.Value] = data.ValueMappingResult{Text: v.Text}
			mappings = append(mappings, m)
		case v.GetFrom() >= 0 || v.GetTo() > 0:
			m := data.RangeValueMapper{From: (*data.ConfFloat64)(&v.From), To: (*data.ConfFloat64)(&v.To), Result: data.ValueMappingResult{Text: v.Text}}
			mappings = append(mappings, m)
		default:
			continue
		}
	}
	if len(mappings) > 0 {
		if cfg == nil {
			cfg = &data.FieldConfig{}
		}
		cfg.Mappings = mappings

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

		frame.Meta = convertFrameMeta(metricFrame.Meta)

		res = append(res, frame)
	}

	// add metadata -> add next token to the first frame (this is how other datasource plugins are doing this)
	if len(res) > 0 {
		frame := res[0]
		if frame.Meta == nil {
			frame.Meta = &data.FrameMeta{
				Custom: models.Metadata{
					NextToken: response.GetNextToken(),
				},
			}
		} else {
			frame.Meta.Custom = models.Metadata{
				NextToken: response.GetNextToken(),
			}
		}
	}

	return res
}

func convertFrameMeta(meta *pb.FrameMeta) *data.FrameMeta {
	if meta == nil {
		return nil
	}
	return &data.FrameMeta{
		Type:                   convertFrameMetaFrameType(meta.Type),
		Custom:                 nil,
		Stats:                  nil,
		Notices:                convertFrameMetaNotices(meta.Notices),
		Channel:                "",
		PreferredVisualization: converFrameMetaVisType(meta.PreferredVisualization),
		ExecutedQueryString:    meta.ExecutedQueryString,
	}
}

func converFrameMetaVisType(t pb.FrameMeta_VisType) data.VisType {
	switch t {
	case pb.FrameMeta_VisTypeTable:
		return data.VisTypeTable
	case pb.FrameMeta_VisTypeLogs:
		return data.VisTypeLogs
	case pb.FrameMeta_VisTypeTrace:
		return data.VisTypeTrace
	case pb.FrameMeta_VisTypeNodeGraph:
		return data.VisTypeNodeGraph
	default:
		return data.VisTypeGraph
	}
}

func convertFrameMetaNotices(notices []*pb.FrameMeta_Notice) []data.Notice {
	var res = make([]data.Notice, len(notices))
	for i := range notices {
		n := notices[i]
		res[i] = convertFrameMetaNotice(n)
	}
	return res
}

func convertFrameMetaNotice(n *pb.FrameMeta_Notice) data.Notice {
	return data.Notice{
		Severity: convertFrameMetaNoticeSeverity(n.Severity),
		Text:     n.Text,
		Link:     n.Link,
		Inspect:  convertFrameMetaNoticeInspectType(n.Inspect),
	}
}

func convertFrameMetaNoticeSeverity(s pb.FrameMeta_Notice_NoticeSeverity) data.NoticeSeverity {
	switch s {
	case pb.FrameMeta_Notice_NoticeSeverityError:
		return data.NoticeSeverityError
	case pb.FrameMeta_Notice_NoticeSeverityWarning:
		return data.NoticeSeverityWarning
	default:
		return data.NoticeSeverityInfo
	}
}

func convertFrameMetaNoticeInspectType(v pb.FrameMeta_Notice_InspectType) data.InspectType {
	switch v {
	case pb.FrameMeta_Notice_InspectTypeMeta:
		return data.InspectTypeMeta
	case pb.FrameMeta_Notice_InspectTypeError:
		return data.InspectTypeError
	case pb.FrameMeta_Notice_InspectTypeData:
		return data.InspectTypeData
	case pb.FrameMeta_Notice_InspectTypeStats:
		return data.InspectTypeStats
	default:
		return data.InspectTypeNone
	}
}

func convertFrameMetaFrameType(t pb.FrameMeta_FrameType) data.FrameType {
	switch t {
	case pb.FrameMeta_FrameTypeTimeSeriesWide:
		return data.FrameTypeTimeSeriesWide
	case pb.FrameMeta_FrameTypeTimeSeriesLong:
		return data.FrameTypeTimeSeriesLong
	case pb.FrameMeta_FrameTypeTimeSeriesMany:
		return data.FrameTypeTimeSeriesMany
	case pb.FrameMeta_FrameTypeTable:
		return data.FrameTypeTable
	case pb.FrameMeta_FrameTypeDirectoryListing:
		return data.FrameTypeDirectoryListing
	default:
		return data.FrameTypeUnknown
	}
}
