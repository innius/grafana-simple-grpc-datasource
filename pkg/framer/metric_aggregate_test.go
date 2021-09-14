package framer

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMetricAggregate_Frames(t *testing.T) {
	ts := time.Date(2021, 9, 13, 10, 29, 00, 00, time.Local)
	r := &pb.GetMetricAggregateResponse{
		Result: []*pb.GetMetricAggregateResponse_Result{
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

	sut := MetricAggregate{
		GetMetricAggregateResponse: r,
		AggregationType:            pb.AggregateType_AVERAGE,
	}

	frames, err := sut.Frames()
	assert.NoError(t, err)
	require.Len(t, frames, 2)
	frame := frames[0]
	assert.Equal(t, "avg", frame.Fields[1].Name)
	meta := frames[0].Meta
	assert.NotNil(t, meta)
	assert.NotNil(t, meta.Custom)
	if assert.IsType(t, models.Metadata{}, meta.Custom) {
		custom := meta.Custom.(models.Metadata)
		assert.NotEmpty(t, custom.NextToken)
	}
}
