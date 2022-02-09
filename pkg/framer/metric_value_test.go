package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestMetricValue_Frames(t *testing.T) {
	ts := time.Date(2022, 01, 19, 16, 03, 10, 00, time.Local)

	sut := &MetricValue{
		GetMetricValueResponse: &pb.GetMetricValueResponse{
			Frames: []*pb.GetMetricValueResponse_Frame{
				{
					Metric: "foo",
					Fields: []*pb.SingleValueField{
						{
							Name: "field_1",
							Labels: []*pb.Label{
								{
									Key:   "zone",
									Value: "a",
								},
							},
							Config: nil,
							Value:  10,
						},
					},
					Timestamp: timestamppb.New(ts),
				},
				{
					Metric: "bar",
					Fields: []*pb.SingleValueField{
						{
							Value: 20,
						},
					},
					Timestamp: timestamppb.New(ts),
				},
			},
		},
		Query: models.MetricValueQuery{
			MetricBaseQuery: models.MetricBaseQuery{
				Dimensions: []models.Dimension{
					{
						Key:   "machine",
						Value: "m1",
					},
				},
				DisplayName: `{{machine}}-{{metric}}-{{zone}}-{{field}}`,
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

	t.Run("the format name expression should be applied", func(t *testing.T) {
		assert.Equal(t, "m1-foo-a-field_1", res[0].Fields[1].Config.DisplayNameFromDS)
	})
}
