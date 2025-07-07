package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	backendapi "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend"

	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/pkg/errors"
)

type Datasource struct {
	backendAPI backendapi.Backend
	queryMux   *datasource.QueryTypeMux
	backend.CallResourceHandler
}

// Make sure SampleDatasource implements required interfaces.
// This is important to do since otherwise we will only get a
// not implemented error response from plugin in runtime.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
	_ backend.StreamHandler         = (*Datasource)(nil)
)

// QueryHandlerFunc is the function signature used for mux.HandleFunc
// Looks like mux.HandleFunc uses backend.QueryHandlerFunc
// type QueryDataHandlerFunc func(ctx context.Context, req *QueryDataRequest) (*QueryDataResponse, error)
type QueryHandlerFunc func(context.Context, backend.QueryDataRequest, backend.DataQuery) backend.DataResponse

func DataResponseErrorUnmarshal(err error) backend.DataResponse {
	return backend.DataResponse{
		Error: errors.Wrap(err, "failed to unmarshal JSON request into query"),
	}
}

func DataResponseErrorRequestFailed(err error) backend.DataResponse {
	return backend.DataResponse{
		Error: err,
	}
}

// GetQueryHandlers creates the QueryTypeMux type for handling queries
func (d *Datasource) registerQueryHandlers() {
	mux := datasource.NewQueryTypeMux()

	mux.HandleFunc(models.QueryMetricValue, d.HandleGetMetricValueQuery)
	mux.HandleFunc(models.QueryMetricHistory, d.HandleGetMetricHistoryQuery)
	mux.HandleFunc(models.QueryMetricAggregate, d.HandleGetMetricAggregate)

	d.queryMux = mux
}

func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	backendAPI, err := backendapi.New(settings)
	if err != nil {
		return nil, err
	}

	return newDatasourceWithBackendAPI(backendAPI)
}

func newDatasourceWithBackendAPI(backendAPI backendapi.Backend) (instancemgmt.Instance, error) {
	srvr := &Datasource{
		backendAPI: backendAPI,
	}
	mux := http.NewServeMux()
	srvr.registerRoutes(mux)
	srvr.CallResourceHandler = httpadapter.New(mux)
	srvr.registerQueryHandlers() // init once
	return srvr, nil
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return d.queryMux.QueryData(ctx, req)
}

func (d *Datasource) SubscribeStream(context.Context, *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	backend.Logger.Info("subscribe")
	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusOK,
	}, nil
}

func (d *Datasource) PublishStream(context.Context, *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

type Q struct {
	QueryType     string `json:"queryType"`
	Range         backend.TimeRange
	IntervalMS    int64 `json:"intervalMs"`
	MaxDataPoints int64 `json:"maxDataPoints"`
	models.MetricBaseQuery
}

func (d *Datasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	// see:https://grafana.com/developers/plugin-tools/tutorials/build-a-streaming-data-source-plugin
	backend.Logger.Info("subscribe")

	backend.Logger.Info("RunStream", "req.Data", string(req.Data))
	query := Q{}
	if err := json.Unmarshal(req.Data, &query); err != nil {
		backend.Logger.Error("Unmarshal error", "error", err)
		return err
	}
	backend.Logger.Info("Q", "Q", query)

	var queryFunc func(backend.TimeRange) (data.Frames, error)

	switch query.QueryType {
	case models.QueryMetricAggregate:
		queryFunc = func(tr backend.TimeRange) (data.Frames, error) {
			q := models.MetricAggregateQuery{
				MetricBaseQuery: query.MetricBaseQuery,
			}
			q.TimeRange = tr
			return d.backendAPI.HandleGetMetricAggregateQuery(ctx, &q)
		}
	case models.QueryMetricHistory:
		queryFunc = func(tr backend.TimeRange) (data.Frames, error) {
			q := models.MetricHistoryQuery{
				MetricBaseQuery: query.MetricBaseQuery,
			}
			q.TimeRange = tr
			return d.backendAPI.HandleGetMetricHistoryQuery(ctx, &q)
		}
	case models.QueryMetricValue:
		queryFunc = func(tr backend.TimeRange) (data.Frames, error) {
			q := models.MetricValueQuery{
				MetricBaseQuery: query.MetricBaseQuery,
			}
			q.TimeRange = tr
			return d.backendAPI.HandleGetMetricValueQuery(ctx, &q)
		}
	default:
		return errors.New("unsupported query type")
	}

	now := time.Now()
	tr := backend.TimeRange{
		From: now.Add(-1 * time.Hour),
		To:   now,
	}

	f, err := queryFunc(tr)
	if err != nil {
		backend.Logger.Error("Query failure", "error", err)
	}
	for _, fr := range f {
		err := sender.SendFrame(fr, data.IncludeAll)
		if err != nil {
			backend.Logger.Error("Failed send frame", "error", err)
		}
	}

	ticker := time.NewTicker(10 * time.Second)

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			backend.Logger.Error("context canceled")
			return ctx.Err()
		case tickTime := <-ticker.C:
			newTr := backend.TimeRange{
				From: tr.To,
				To:   tickTime,
			}
			// On first tick, if tr.To is zero time, use original From as From
			if tr.To.IsZero() {
				newTr.From = tr.From
			}
			tr = newTr
			f, err := queryFunc(tr)
			if err != nil {
				backend.Logger.Error("Query failure", "error", err)
			}
			for _, fr := range f {
				err := sender.SendFrame(fr, data.IncludeAll)
				if err != nil {
					backend.Logger.Error("Failed send frame", "error", err)
				}
			}
		}
	}
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	_, err := d.backendAPI.GetDimensionKeys(ctx, models.GetDimensionKeysRequest{})
	if err != nil {
		switch status.Code(err) {
		case codes.Unauthenticated:
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: "authentication error; please check if your datasource is provided with valid credentials",
			}, nil
		case codes.Unavailable:
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: "could not establish a connection; please check if your datasource is provided with valid credentials",
			}, nil
		default:
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: err.Error(),
			}, nil
		}
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: backend.HealthStatusOk.String(),
	}, nil
}

func (d *Datasource) Dispose() {
	d.backendAPI.Dispose()
}
