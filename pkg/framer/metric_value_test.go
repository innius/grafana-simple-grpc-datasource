package framer

import (
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMetricValue_Frames(t *testing.T) {
	ts := time.Date(2021, 9, 13, 10, 29, 00, 00, time.Local)

	r := &pb.GetMetricValueResponse{
		Result: []*pb.GetMetricValueResponse_Result{
			{
				Metric: &pb.Metric{Id: "foo"},
				Datapoint: &pb.Datapoint{
					Timestamp: ts.Unix(),
					Value: &pb.MetricValue{
						DoubleValue: 19.75,
					},
				},
			},
			{
				Metric: &pb.Metric{Id: "bar"},
				Datapoint: &pb.Datapoint{
					Timestamp: ts.Unix(),
					Value: &pb.MetricValue{
						DoubleValue: 19.74,
					},
				},
			},
		},
	}

	sut := MetricValue{r}

	frames, err := sut.Frames()
	assert.NoError(t, err)
	assert.Len(t, frames, 2)
	f1 := frames[0]
	assert.Equal(t, "foo", f1.Name)
	require.Len(t, f1.Fields, 2)
	assert.Equal(t, "time", f1.Fields[0].Name)
	assert.Equal(t, "Value", f1.Fields[1].Name)
	assert.Equal(t, ts,  f1.Fields[0].At(0))
	assert.Equal(t, 19.75, f1.Fields[1].At(0))
}
