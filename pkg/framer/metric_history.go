package framer

import (
	"strings"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type MetricHistory struct {
	*pb.GetMetricHistoryResponse
	Query models.MetricHistoryQuery
}

func (f MetricHistory) FormatDisplayName(frame *pb.Frame, fld *pb.Field) string {
	var args []Arg
	for key, value := range f.Query.Options {
		args = append(args, Arg{Key: strings.ToLower(key), Value: value.Label})
	}
	return formatDisplayName(FormatDisplayNameInput{
		DisplayName: f.Query.DisplayName,
		FieldName:   fld.Name,
		MetricID:    frame.Metric,
		Dimensions:  f.Query.Dimensions,
		Labels:      fld.GetLabels(),
		Args:        args,
	})
}

func (f MetricHistory) Frames() (data.Frames, error) {
	return convertToDataFrames(f), nil
}
