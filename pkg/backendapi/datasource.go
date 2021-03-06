package backendapi

import (
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/client"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backendapi/connector"
	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"google.golang.org/grpc"
)

type Datasource struct {
	client client.BackendAPIClient
	conn   *grpc.ClientConn
}

func NewDatasource(settings backend.DataSourceInstanceSettings) (*Datasource, error) {
	cfg := client.BackendAPIDatasourceSettings{}
	err := cfg.Load(settings)
	if err != nil {
		return nil, err
	}
	cl, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Datasource{
		client: cl,
	}, nil
}

func (ds *Datasource) HandleGetMetricValueQuery(ctx context.Context, query *models.MetricValueQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.GetMetricValue(ctx, ds.client, *query)
	if err != nil {
		return nil, err
	}
	return res.Frames()
}

func (ds *Datasource) HandleGetMetricHistoryQuery(ctx context.Context, query *models.MetricHistoryQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.GetMetricHistory(ctx, ds.client, *query)
	if err != nil {
		return backendErrorResponse(err)
	}
	return res.Frames()
}

func (ds *Datasource) HandleGetMetricAggregateQuery(ctx context.Context, query *models.MetricAggregateQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.GetMetricAggregate(ctx, ds.client, *query)
	if err != nil {
		return backendErrorResponse(err)
	}
	return res.Frames()
}

func (ds *Datasource) HandleListDimensionsQuery(ctx context.Context, query *models.DimensionKeysQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.ListDimensionKeys(ctx, ds.client, *query)
	if err != nil {
		return backendErrorResponse(err)
	}
	return res.Frames()
}

func (ds *Datasource) HandleListDimensionValuesQuery(ctx context.Context, query *models.DimensionValueQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.ListDimensionValues(ctx, ds.client, *query)
	if err != nil {
		return backendErrorResponse(err)
	}
	return res.Frames()
}

func (ds *Datasource) HandleListMetricsQuery(ctx context.Context, query *models.MetricsQuery) (data.Frames, error) {
	//TODO: remove pointer dereference
	res, err := connector.ListMetrics(ctx, ds.client, *query)
	if err != nil {
		return backendErrorResponse(err)
	}
	return res.Frames()
}

func (ds *Datasource) Dispose() {
	ds.client.Dispose()
}
