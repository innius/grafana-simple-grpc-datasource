package framer

import (
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"testing"
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

	assert.Equal(t, data.FrameTypeTimeSeriesMany, string(res.Type))

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
