package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMetricValue_Frames(t *testing.T) {
	ts := time.Date(2022, 01, 19, 16, 03, 10, 00, time.Local)

	sut := &MetricValue{
		GetMetricValueResponse: pb.GetMetricValueResponse{
			Result: []*pb.GetMetricValueResponse_Result{
				{
					Metric:    "foo",
					Timestamp: ts.Unix(),
					Values: []*pb.GetMetricValueResponse_Result_MetricValue{
						{
							DoubleValue: 1,
							Id:          "a",
						},
					},
				},
				{
					Metric:    "bar",
					Timestamp: ts.Unix(),
					Values: []*pb.GetMetricValueResponse_Result_MetricValue{
						{
							DoubleValue: 2,
							Id:          "b",
						},
					},
				},
			},
		},
		MetricValueQuery: models.MetricValueQuery{},
	}

	res, err := sut.Frames()

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "foo", res[0].Name)
}
