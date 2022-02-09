package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricHistory struct {
	*pb.GetMetricHistoryResponse
	Query models.MetricHistoryQuery
}

func (f MetricHistory) FormatDisplayName(frame *pb.Frame, fld *pb.Field) string {
	return formatDisplayName(FormatDisplayNameInput{
		DisplayName: f.Query.DisplayName,
		FieldName:   fld.Name,
		MetricID:    frame.Metric,
		Dimensions:  f.Query.Dimensions,
		Labels:      fld.GetLabels(),
	})
}

func (f MetricHistory) Frames() (data.Frames, error) {
	return convertToDataFrames(f), nil
}
