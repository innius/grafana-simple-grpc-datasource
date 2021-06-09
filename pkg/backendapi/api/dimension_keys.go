package api

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/framer"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto"
	"context"
)

func ListDimensionKeys(ctx context.Context, client client.BackendAPIClient, query models.DimensionKeysQuery) (*framer.DimensionKeys, error) {
	resp, err := client.ListDimensionKeys(ctx, &pb.ListDimensionKeysRequest{
		Filter: query.Filter,
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
