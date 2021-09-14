package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMetricHistory_Frames(t *testing.T) {
	ts := time.Date(2021, 9, 13, 10, 29, 00, 00, time.Local)
	r := &pb.GetMetricHistoryResponse{
		Result: []*pb.GetMetricHistoryResponse_Result{
			{
				Metric: &pb.Metric{
					Id: "foo",
				},
				Values: []*pb.Datapoint{
					{
						Timestamp: ts.Unix(),
						Value: &pb.MetricValue{
							DoubleValue: 19.75,
						},
					},
				},
			},
			{
				Metric: &pb.Metric{
					Id: "bar",
				},
				Values: []*pb.Datapoint{
					{
						Timestamp: ts.Unix(),
						Value: &pb.MetricValue{
							DoubleValue: 19.74,
						},
					},
				},
			},
		},
		NextToken: "next",
	}

	sut := MetricHistory{r}

	frames, err := sut.Frames()
	assert.NoError(t, err)
	assert.Len(t, frames, 2)
	meta := frames[0].Meta
	assert.NotNil(t, meta)
	assert.NotNil(t, meta.Custom)
	if assert.IsType(t, models.Metadata{}, meta.Custom) {
		custom := meta.Custom.(models.Metadata)
		assert.NotEmpty(t, custom.NextToken)
	}
}
