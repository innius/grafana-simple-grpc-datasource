package connector

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
)

func ListDimensionValues(ctx context.Context, client client.BackendAPIClient, query models.GetDimensionValuesRequest) (*models.GetDimensionValueResponse, error) {
	resp, err := client.ListDimensionValues(ctx, &pb.ListDimensionValuesRequest{
		DimensionKey: query.DimensionKey,
		Filter:       query.Filter,
		SelectedDimensions: Map(query.SelectedDimensions, func(dimension models.Dimension) *pb.Dimension {
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
		Values: Map(resp.GetResults(), func(dimension *pb.ListDimensionValuesResponse_Result) models.DimensionValueDefinition {
			return models.DimensionValueDefinition{
				Value:       dimension.Value,
				Label:       dimension.Value,
				Description: dimension.Description,
			}
		}),
	}, nil
}
