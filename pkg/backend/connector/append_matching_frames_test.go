package connector

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
)

func TestAppendMatchingFrames(t *testing.T) {
	res := []*pb.Frame{
		{
			Metric: "temperature", Fields: []*pb.Field{
				{Name: "value", Values: []float64{1, 2, 3}},
			},
			Timestamps: []*timestamppb.Timestamp{timestamppb.New(time.Now())},
		},
	}
	frames := map[string]*pb.Frame{}

	t.Run("append the first frame", func(t *testing.T) {
		appendMatchingFrames(frames, res)

		assert.Equal(t, map[string]*pb.Frame{"temperature": res[0]}, frames)
	})
	t.Run("append the second frame", func(t *testing.T) {
		appendMatchingFrames(frames, res)
		exp := &pb.Frame{
			Metric: "temperature",
			Timestamps: []*timestamppb.Timestamp{
				res[0].Timestamps[0], res[0].Timestamps[0],
			},
			Fields: []*pb.Field{res[0].Fields[0], res[0].Fields[0]},
		}
		assert.Equal(t, exp.Timestamps, frames["temperature"].Timestamps)
		assert.Equal(t, exp.Fields[0].Values, frames["temperature"].Fields[0].Values)
	})
}
