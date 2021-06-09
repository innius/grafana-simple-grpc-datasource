package server

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Datasource interface {
	HandleGetMetricValueQuery(ctx context.Context, query *models.MetricValueQuery) (data.Frames, error)
	HandleGetMetricHistoryQuery(ctx context.Context, query *models.MetricHistoryQuery) (data.Frames, error)
	HandleListDimensionsQuery(ctx context.Context, query *models.DimensionKeysQuery) (data.Frames, error)
	HandleListDimensionValuesQuery(ctx context.Context, query *models.DimensionValueQuery) (data.Frames, error)
	HandleListMetricsQuery(ctx context.Context, query *models.MetricsQuery) (data.Frames, error)
	Dispose()
}
