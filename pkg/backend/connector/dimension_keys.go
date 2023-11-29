package connector

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/samber/lo"
)

func ListDimensionKeys(ctx context.Context, client client.BackendAPIClient, query models.GetDimensionKeysRequest) (*models.GetDimensionKeysResponse, error) {
	resp, err := client.ListDimensionKeys(ctx, &pb.ListDimensionKeysRequest{
		Filter: query.Filter,
		SelectedDimensions: lo.Map(query.SelectedDimensions, func(dimension models.Dimension, _ int) *pb.Dimension {
			return &pb.Dimension{
				Key:   dimension.Key,
				Value: dimension.Value,
			}
		}),
	})

	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	return &models.GetDimensionKeysResponse{
		Keys: lo.Map(resp.GetResults(), func(dimension *pb.ListDimensionKeysResponse_Result, _ int) models.DimensionKeyDefinition {
			return models.DimensionKeyDefinition{
				Value:       dimension.Key,
				Label:       dimension.Key,
				Description: dimension.Description,
			}
		}),
	}, nil
}
