package backend

import (
	"context"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend/connector"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"google.golang.org/grpc"
)

type Backend interface {
	HandleGetMetricValueQuery(ctx context.Context, query *models.MetricValueQuery) (data.Frames, error)
	HandleGetMetricHistoryQuery(ctx context.Context, query *models.MetricHistoryQuery) (data.Frames, error)
	HandleGetMetricAggregateQuery(ctx context.Context, query *models.MetricAggregateQuery) (data.Frames, error)

	GetDimensionKeys(ctx context.Context, query models.GetDimensionKeysRequest) (*models.GetDimensionKeysResponse, error)
	GetDimensionValues(ctx context.Context, query models.GetDimensionValuesRequest) (*models.GetDimensionValueResponse, error)
	GetMetrics(ctx context.Context, query models.GetMetricsRequest) (*models.GetMetricsResponse, error)
	GetQueryOptions(ctx context.Context, input models.GetQueryOptionsRequest) (*models.GetQueryOptionsResponse, error)

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

func (ds *backendImpl) GetDimensionKeys(ctx context.Context, query models.GetDimensionKeysRequest) (*models.GetDimensionKeysResponse, error) {
	res, err := connector.ListDimensionKeys(ctx, ds.client, query)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ds *backendImpl) GetDimensionValues(ctx context.Context, query models.GetDimensionValuesRequest) (*models.GetDimensionValueResponse, error) {
	res, err := connector.ListDimensionValues(ctx, ds.client, query)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (ds *backendImpl) GetMetrics(ctx context.Context, query models.GetMetricsRequest) (*models.GetMetricsResponse, error) {
	res, err := connector.ListMetrics(ctx, ds.client, query)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (backendimpl *backendImpl) GetQueryOptions(ctx context.Context, input models.GetQueryOptionsRequest) (*models.GetQueryOptionsResponse, error) {
	return connector.GetQueryOptionDefinitions(ctx, backendimpl.client, input)
}

func (ds *backendImpl) Dispose() {
	ds.client.Dispose()
}
