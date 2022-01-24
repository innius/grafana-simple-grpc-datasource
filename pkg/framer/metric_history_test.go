package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMetricHistory_Frames(t *testing.T) {
	ts := time.Date(2022, 01, 19, 16, 03, 10, 00, time.Local)

	sut := MetricHistory{
		GetMetricHistoryResponse: pb.GetMetricHistoryResponse{
			Data: []*pb.GetMetricHistoryResponse_Data{
				{
					Metric: "foo",
					Series: []*pb.GetMetricHistoryResponse_Data_TimeSeries{
						{
							Timestamp: ts.Unix(),
							Values: []*pb.GetMetricHistoryResponse_Data_TimeSeries_MetricValue{
								{
									DoubleValue: 10,
									Id:          "a",
								},
								{
									DoubleValue: 20,
									Id:          "b",
								},
							},
						},
					},
				},
				{
					Metric: "bar",
					Series: []*pb.GetMetricHistoryResponse_Data_TimeSeries{
						{
							Timestamp: ts.Unix(),
							Values: []*pb.GetMetricHistoryResponse_Data_TimeSeries_MetricValue{
								{
									DoubleValue: 30,
									Id:          "c",
								},
								{
									DoubleValue: 40,
									Id:          "d",
								},
							},
						},
					},
				},
			},
			NextToken: "next-please",
		},
		MetricHistoryQuery: models.MetricHistoryQuery{},
	}

	res, err := sut.Frames()

	assert.NoError(t, err)
	assert.NotNil(t, res)
	if assert.Len(t, res, 2) {
		metricOne := res[0]
		assert.Equal(t, "foo", metricOne.Name)
		assert.Len(t, metricOne.Fields, 3)
		for _, f := range metricOne.Fields {
			assert.Contains(t, []string{"time", "a", "b"}, f.Name)
		}
	}

	assert.Equal(t, "next-please", res[0].Meta.Custom.(models.Metadata).NextToken)
}
