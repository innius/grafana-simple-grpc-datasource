package connector

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v2"
	"context"
)

func ListDimensionValues(ctx context.Context, client client.BackendAPIClient, query models.DimensionValueQuery) (*framer.DimensionValues, error) {
	resp, err := client.ListDimensionValues(ctx, &pb.ListDimensionValuesRequest{
		DimensionKey: query.DimensionKey,
		Filter:       query.Filter,
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
