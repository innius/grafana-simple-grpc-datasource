package connector

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/samber/lo"
)

func ListDimensionValues(ctx context.Context, client client.BackendAPIClient, query models.GetDimensionValuesRequest) (*models.GetDimensionValueResponse, error) {
	resp, err := client.ListDimensionValues(ctx, &pb.ListDimensionValuesRequest{
		DimensionKey: query.DimensionKey,
		Filter:       query.Filter,
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

	return &models.GetDimensionValueResponse{
		Values: lo.Map(resp.GetResults(), func(dimension *pb.ListDimensionValuesResponse_Result, _ int) models.DimensionValueDefinition {
			return models.DimensionValueDefinition{
				Value:       dimension.Value,
				Label:       dimension.Value,
				Description: dimension.Description,
			}
		}),
	}, nil
}
