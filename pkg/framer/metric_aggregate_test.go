package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMetricAggregate_Frames(t *testing.T) {
	ts := time.Date(2022, 01, 19, 16, 03, 10, 00, time.Local)

	sut := MetricAggregate{
		GetMetricAggregateResponse: &pb.GetMetricAggregateResponse{
			Data: []*pb.GetMetricAggregateResponse_Data{
				{
					Metric: &pb.Metric{
						Id: "foo",
					},
					Labels: []*pb.Label{
						{
							Key:   "zone",
							Value: "a",
						},
					},
					Series: []*pb.GetMetricAggregateResponse_Data_TimeSeries{
						{
							Timestamp: ts.Unix(),
							Value: &pb.GetMetricAggregateResponse_Data_TimeSeries_MetricValue{
								DoubleValue: 10,
							},
						},
					},
				},
				{
					Metric: &pb.Metric{
						Id: "bar",
					},
					Series: []*pb.GetMetricAggregateResponse_Data_TimeSeries{
						{
							Timestamp: ts.Unix(),
							Value: &pb.GetMetricAggregateResponse_Data_TimeSeries_MetricValue{
								DoubleValue: 10,
							},
						},
					},
				},
			},
			NextToken: "next-please",
		},
		Query: models.MetricBaseQuery{
			Dimensions: []models.Dimension{
				{
					Key:   "machine",
					Value: "m1",
				},
			},
			DisplayName: `{{machine}}-{{metric}}-{{zone}}-{{aggregate}}`,
		},
		AggregateType: pb.AggregateType_AVERAGE,
	}

	res, err := sut.Frames()

	assert.NoError(t, err)
	t.Run("the result should contain two frames", func(t *testing.T) {
		assert.Len(t, res, 2)
	})
	t.Run("the data frame should have a name", func(t *testing.T) {
		assert.Equal(t, "foo", res[0].Name)
	})

	t.Run("the format name expression should be applied", func(t *testing.T) {
		assert.Equal(t, "m1-foo-a-avg", res[0].Fields[1].Config.DisplayNameFromDS)
	})

	t.Run("the data frame should have a NextToken", func(t *testing.T) {
		assert.Equal(t, "next-please", res[0].Meta.Custom.(models.Metadata).NextToken)
	})
}
