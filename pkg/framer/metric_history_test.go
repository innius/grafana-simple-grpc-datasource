package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMetricHistory_Frames(t *testing.T) {
	ts := time.Date(2022, 01, 19, 16, 03, 10, 00, time.Local)

	sut := MetricHistory{
		GetMetricHistoryResponse: &pb.GetMetricHistoryResponse{
			Data: []*pb.GetMetricHistoryResponse_Data{
				{
					Metric: &pb.Metric{
						Id:   "foo",
						Unit: "℃",
					},
					Labels: []*pb.Label{
						{
							Key:   "zone",
							Value: "a",
						},
					},
					Series: []*pb.GetMetricHistoryResponse_Data_TimeSeries{
						{
							Timestamp: ts.Unix(),
							Value: &pb.GetMetricHistoryResponse_Data_TimeSeries_MetricValue{
								DoubleValue: 10,
							},
						},
					},
				},
				{
					Metric: &pb.Metric{
						Id: "bar",
					},
					Series: []*pb.GetMetricHistoryResponse_Data_TimeSeries{
						{
							Timestamp: ts.Unix(),
							Value: &pb.GetMetricHistoryResponse_Data_TimeSeries_MetricValue{
								DoubleValue: 20,
							},
						},
					},
				},
			},
			NextToken: "next-please",
		},
		MetricHistoryQuery: models.MetricHistoryQuery{
			MetricBaseQuery: models.MetricBaseQuery{
				Dimensions: []models.Dimension{
					{
						Key:   "machine",
						Value: "m1",
					},
				},
				DisplayName: `{{machine}}-{{metric}}-{{zone}}`,
			},
		},
	}

	res, err := sut.Frames()

	assert.NoError(t, err)

	t.Run("the result should contain two frames", func(t *testing.T) {
		assert.Len(t, res, 2)
	})
	t.Run("the data frame should have a name", func(t *testing.T) {
		assert.Equal(t, "foo", res[0].Name)
	})
	t.Run("the data field", func(t *testing.T) {
		dataField := res[0].Fields[1]
		t.Run("the format name expression should be applied", func(t *testing.T) {
			assert.Equal(t, "m1-foo-a", dataField.Config.DisplayNameFromDS)
		})
		t.Run("the data field should have the expected unit", func(t *testing.T) {
			assert.Equal(t, "℃", dataField.Config.Unit)

		})
		t.Run("should have labels", func(t *testing.T) {
			assert.Equal(t, data.Labels{"zone": "a"}, dataField.Labels)
		})
	})

	t.Run("the data frame should have a NextToken", func(t *testing.T) {
		assert.Equal(t, "next-please", res[0].Meta.Custom.(models.Metadata).NextToken)
	})

}
