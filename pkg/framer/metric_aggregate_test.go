package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMetricAggregate_Frames(t *testing.T) {
	sut := MetricAggregate{
		MetricAggregateQuery: models.MetricAggregateQuery{
			MetricBaseQuery: models.MetricBaseQuery{},
			AggregateType:   "",
		},
		GetMetricAggregateResponse: pb.GetMetricAggregateResponse{
			Values: []*pb.MetricHistoryValue{
				{
					Timestamp: time.Date(2021, 01, 19, 14, 55, 01, 00, time.Local).Unix(),
					Values: []*pb.MetricValue{
						{
							DoubleValue: 25.04,
							Id:          "avg",
						},
					},
				},
			},
			NextToken: "",
		},
		AggregationType: 0,
	}

	res, err := sut.Frames()
	assert.NoError(t, err)
	exp := data.Frames{}
	assert.Equal(t, exp, res)
}
