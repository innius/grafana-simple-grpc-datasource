package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"time"
)

func getTime(timeInSeconds int64) time.Time {
	return time.Unix(timeInSeconds, 0)
}

func newDataFieldForMetric(metric *pb.Metric, labels []*pb.Label, displayName string, length int) *data.Field {
	return newDataFieldForMetric2(metric, metric.Id, labels, displayName, length)
}

func newDataFieldForMetric2(metric *pb.Metric, id string, labels []*pb.Label, displayName string, length int) *data.Field {
	dataField := fields.MetricField(id, length)
	dataField.Config = &data.FieldConfig{}
	if displayName != "" {
		dataField.Config.DisplayNameFromDS = displayName
	}
	if metric.Unit != "" {
		dataField.Config.Unit = metric.Unit
	}

	if len(labels) > 0 {
		dataField.Labels = data.Labels{}
		for i := range labels {
			label := labels[i]
			dataField.Labels[label.Key] = label.Value
		}
	}
	return dataField
}
