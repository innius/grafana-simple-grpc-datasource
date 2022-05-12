package connector

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"context"
)

func ListDimensionKeys(ctx context.Context, client client.BackendAPIClient, query models.DimensionKeysQuery) (*framer.DimensionKeys, error) {
	selectedDimensions := make([]*pb.Dimension, len(query.SelectedDimensions))
	for _, dimension := range query.SelectedDimensions {
		selectedDimensions = append(selectedDimensions, &pb.Dimension{
			Key:   dimension.Key,
			Value: dimension.Value,
		})
	}
	resp, err := client.ListDimensionKeys(ctx, &pb.ListDimensionKeysRequest{
		Filter:             query.Filter,
		SelectedDimensions: selectedDimensions,
	})

	if err != nil {
		return nil, err
	}
	return &framer.DimensionKeys{
		ListDimensionKeysResponse: pb.ListDimensionKeysResponse{
			Results: resp.Results,
		},
	}, nil
}
