package connector

import (
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/samber/lo"
)

func appendMatchingFrames(frames map[string]*pb.Frame, newFrames []*pb.Frame) {
	for _, frame := range newFrames {
		// Check if the metric already exists in the frames map
		fr, exists := frames[frame.Metric]
		if !exists {
			frames[frame.Metric] = frame
			continue
		}

		// Iterate through each field in the current frame
		for _, fld := range frame.Fields {
			// Check if the field already exists in the existing frame
			if field, found := lo.Find(fr.Fields, func(f *pb.Field) bool { return f.Name == fld.Name }); found {
				field.Values = append(field.Values, fld.Values...)
			} else {
				fr.Fields = append(fr.Fields, fld)
			}
		}

		// Append the timestamps from the current frame
		fr.Timestamps = append(fr.Timestamps, frame.Timestamps...)
	}
}
