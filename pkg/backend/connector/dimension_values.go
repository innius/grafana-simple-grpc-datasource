package connector

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
)

func ListDimensionValues(ctx context.Context, client client.BackendAPIClient, query models.DimensionValueQuery) (*framer.DimensionValues, error) {
	selectedDimensions := make([]*pb.Dimension, len(query.SelectedDimensions))
	for _, dimension := range query.SelectedDimensions {
		selectedDimensions = append(selectedDimensions, &pb.Dimension{
			Key:   dimension.Key,
			Value: dimension.Value,
		})
	}
	resp, err := client.ListDimensionValues(ctx, &pb.ListDimensionValuesRequest{
		DimensionKey:       query.DimensionKey,
		Filter:             query.Filter,
		SelectedDimensions: selectedDimensions,
	})

	if err != nil {
		return nil, err
	}
	return &framer.DimensionValues{
		ListDimensionValuesResponse: pb.ListDimensionValuesResponse{
			Results: resp.Results,
		},
	}, nil
}
