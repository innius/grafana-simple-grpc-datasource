package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer/fields"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Metrics struct {
	pb.ListMetricsResponse
}

func (p Metrics) Frames() (data.Frames, error) {
	length := len(p.Metrics)
	labelField := fields.NewFieldWithName("label", data.FieldTypeString, length)
	nameField := fields.NewFieldWithName("value", data.FieldTypeString, length)
	descriptionField := fields.NewFieldWithName("description", data.FieldTypeString, length)

	frame := data.NewFrame("dimensions", labelField, nameField, descriptionField)

	for i, d := range p.Metrics {
		labelField.Set(i, d.Name)
		nameField.Set(i, d.Name)
		descriptionField.Set(i, d.Description)
	}

	return data.Frames{frame}, nil
}
