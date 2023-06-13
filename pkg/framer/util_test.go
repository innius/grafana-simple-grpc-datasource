package framer

import (
	"reflect"
	"testing"

	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

func TestConvertMetadata(t *testing.T) {
	m := &pb.FrameMeta{
		Type: pb.FrameMeta_FrameTypeTimeSeriesMany,
		Notices: []*pb.FrameMeta_Notice{
			{
				Severity: pb.FrameMeta_Notice_NoticeSeverityWarning,
				Text:     "Warning",
				Link:     "https://foo.bar/warning",
				Inspect:  pb.FrameMeta_Notice_InspectTypeMeta,
			},
			{
				Severity: pb.FrameMeta_Notice_NoticeSeverityError,
				Text:     "Error",
				Link:     "https://foo.bar/error",
				Inspect:  pb.FrameMeta_Notice_InspectTypeMeta,
			},
			{
				Severity: pb.FrameMeta_Notice_NoticeSeverityInfo,
				Text:     "Info",
				Link:     "https://foo.bar/info",
				Inspect:  pb.FrameMeta_Notice_InspectTypeMeta,
			},
		},
		PreferredVisualization: pb.FrameMeta_VisTypeTable,
		ExecutedQueryString:    "select *",
	}

	res := convertFrameMeta(m)

	assert.Equal(t, string(data.FrameTypeTimeSeriesMany), string(res.Type))

	assert.Equal(t, m.ExecutedQueryString, res.ExecutedQueryString)
	assert.Equal(t, data.VisTypeTable, string(res.PreferredVisualization))

	expectedNotices := []data.Notice{
		{
			Severity: data.NoticeSeverityWarning,
			Text:     m.Notices[0].Text,
			Link:     m.Notices[0].Link,
			Inspect:  data.InspectTypeMeta,
		},
		{
			Severity: data.NoticeSeverityError,
			Text:     m.Notices[1].Text,
			Link:     m.Notices[1].Link,
			Inspect:  data.InspectTypeMeta,
		},
		{
			Severity: data.NoticeSeverityInfo,
			Text:     m.Notices[2].Text,
			Link:     m.Notices[2].Link,
			Inspect:  data.InspectTypeMeta,
		},
	}
	assert.Equal(t, expectedNotices, res.Notices)
}

func Test_convertToDataFieldConfig(t *testing.T) {
	type args struct {
		config            *pb.Config
		formatDisplayName string
	}
	from, to := float64(0), float64(10)
	tests := []struct {
		name string
		args args
		want *data.FieldConfig
	}{
		{
			name: "single value mapping",
			args: args{config: &pb.Config{Mappings: []*pb.ValueMapping{{Value: "1", Text: "ON"}}}},
			want: &data.FieldConfig{Mappings: []data.ValueMapping{data.ValueMapper{"1": data.ValueMappingResult{Text: "ON"}}}},
		},
		{
			name: "range mapping",
			args: args{config: &pb.Config{Mappings: []*pb.ValueMapping{{From: 0, To: 10, Text: "ON"}}}},
			want: &data.FieldConfig{Mappings: []data.ValueMapping{data.RangeValueMapper{From: (*data.ConfFloat64)(&from), To: (*data.ConfFloat64)(&to), Result: data.ValueMappingResult{Text: "ON"}}}},
		},
		{
			name: "empty mapping",
			args: args{config: &pb.Config{Mappings: []*pb.ValueMapping{}}},
			want: &data.FieldConfig{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToDataFieldConfig(tt.args.config, tt.args.formatDisplayName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToDataFieldConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
