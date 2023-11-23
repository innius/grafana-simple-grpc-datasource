package connector

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
)

func ListDimensionKeys(ctx context.Context, client client.BackendAPIClient, query models.GetDimensionKeysRequest) (*models.GetDimensionKeysResponse, error) {
	resp, err := client.ListDimensionKeys(ctx, &pb.ListDimensionKeysRequest{
		Filter: query.Filter,
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

	return &models.GetDimensionKeysResponse{
		Keys: Map(resp.GetResults(), func(dimension *pb.ListDimensionKeysResponse_Result) models.DimensionKeyDefinition {
			return models.DimensionKeyDefinition{
				Value:       dimension.Key,
				Label:       dimension.Key,
				Description: dimension.Description,
			}
		}),
	}, nil
}

func Map[T, R any](collection []T, iteratee func(T) R) []R {
	res := make([]R, len(collection))
	for i := range collection {
		res[i] = iteratee(collection[i])
	}
	return res
}
