package connector

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	pb "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/samber/lo"
)

func GetQueryOptions(ctx context.Context, client client.BackendAPIClient, qt pb.GetOptionsRequest_QueryType) (models.Options, error) {
	resp, err := client.GetQueryOptions(ctx, &pb.GetOptionsRequest{
		QueryType: qt,
	})

	if err != nil {
		return nil, err
	}

	options := lo.Map(resp.Options, func(o *v3.Option, _ int) models.Option {
		return models.Option{
			ID:          o.Id,
			Label:       o.Label,
			Description: o.Description,
			Type:        o.Type.String(),
			EnumValues: lo.Map(o.EnumValues, func(v *v3.EnumValue, _ int) models.EnumValue {
				return models.EnumValue{
					Label:       v.Label,
					ID:          v.Id,
					Description: v.Description,
				}
			}),
			Required: o.Required,
		}
	})
	return options, nil
}
