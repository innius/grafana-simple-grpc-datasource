package connector

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/samber/lo"
)

func GetQueryOptionDefinitions(ctx context.Context, client client.BackendAPIClient, input models.GetQueryOptionsRequest) (models.Options, error) {
	var qt v3.GetOptionsRequest_QueryType
	switch input.QueryType {
	case models.QueryMetricValue:
		qt = v3.GetOptionsRequest_GetMetricValue
	case models.QueryMetricHistory:
		qt = v3.GetOptionsRequest_GetMetricHistory
	default:
		qt = v3.GetOptionsRequest_GetMetricAggregate
	}
	resp, err := client.GetQueryOptions(ctx, &v3.GetOptionsRequest{
		QueryType:       qt,
		SelectedOptions: input.SelectedOptions,
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
					Default:     v.Default,
				}
			}),
			Required: o.Required,
		}
	})
	return options, nil
}
