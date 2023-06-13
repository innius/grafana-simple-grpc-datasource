package backend

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/connector"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	v3 "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/proto/v3"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"google.golang.org/grpc"
)

type GetQueryOptionsRequest struct {
	QueryType string
}

type GetQueryOptionsResponse struct {
	Options models.Options
}

type Backend interface {
	HandleGetMetricValueQuery(ctx context.Context, query *models.MetricValueQuery) (data.Frames, error)
	HandleGetMetricHistoryQuery(ctx context.Context, query *models.MetricHistoryQuery) (data.Frames, error)
	HandleGetMetricAggregateQuery(ctx context.Context, query *models.MetricAggregateQuery) (data.Frames, error)
	HandleListDimensionsQuery(ctx context.Context, query *models.DimensionKeysQuery) (data.Frames, error)
	HandleListDimensionValuesQuery(ctx context.Context, query *models.DimensionValueQuery) (data.Frames, error)
	HandleListMetricsQuery(ctx context.Context, query *models.MetricsQuery) (data.Frames, error)
	GetQueryOptions(ctx context.Context, input GetQueryOptionsRequest) (*GetQueryOptionsResponse, error)
	Dispose()
}

type backendImpl struct {
	client client.BackendAPIClient
	conn   *grpc.ClientConn
}

func New(settings backend.DataSourceInstanceSettings) (Backend, error) {
	cfg := client.BackendAPIDatasourceSettings{}
	err := cfg.Load(settings)
	if err != nil {
		return nil, err
	}
	cl, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	return &backendImpl{
		client: cl,
	}, nil
}

func (ds *backendImpl) HandleGetMetricValueQuery(ctx context.Context, query *models.MetricValueQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.GetMetricValue(ctx, ds.client, *query)
	if err != nil {
		return nil, err
	}
	return res.Frames()
}

func (ds *backendImpl) HandleGetMetricHistoryQuery(ctx context.Context, query *models.MetricHistoryQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.GetMetricHistory(ctx, ds.client, *query)
	if err != nil {
		return backendErrorResponse(err)
	}
	return res.Frames()
}

func (ds *backendImpl) HandleGetMetricAggregateQuery(ctx context.Context, query *models.MetricAggregateQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.GetMetricAggregate(ctx, ds.client, *query)
	if err != nil {
		return backendErrorResponse(err)
	}
	return res.Frames()
}

func (ds *backendImpl) HandleListDimensionsQuery(ctx context.Context, query *models.DimensionKeysQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.ListDimensionKeys(ctx, ds.client, *query)
	if err != nil {
		return backendErrorResponse(err)
	}
	return res.Frames()
}

func (ds *backendImpl) HandleListDimensionValuesQuery(ctx context.Context, query *models.DimensionValueQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.ListDimensionValues(ctx, ds.client, *query)
	if err != nil {
		return backendErrorResponse(err)
	}
	return res.Frames()
}

func (ds *backendImpl) HandleListMetricsQuery(ctx context.Context, query *models.MetricsQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.ListMetrics(ctx, ds.client, *query)
	if err != nil {
		return backendErrorResponse(err)
	}
	return res.Frames()
}

func (backendimpl *backendImpl) GetQueryOptions(ctx context.Context, input GetQueryOptionsRequest) (*GetQueryOptionsResponse, error) {
	var qt v3.GetOptionsRequest_QueryType
	switch input.QueryType {
	case models.QueryMetricValue:
		qt = v3.GetOptionsRequest_GetMetricValue
	case models.QueryMetricHistory:
		qt = v3.GetOptionsRequest_GetMetricHistory
	default:
		qt = v3.GetOptionsRequest_GetMetricAggregate
	}
	res, err := connector.GetQueryOptions(ctx, backendimpl.client, qt)
	if err != nil {
		return nil, err
	}
	return &GetQueryOptionsResponse{Options: res}, nil
}

func (ds *backendImpl) Dispose() {
	ds.client.Dispose()
}
