package plugin

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	backendapi "bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/backend"

	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/data"
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
func (ds *Datasource) registerQueryHandlers() {
	mux := datasource.NewQueryTypeMux()

	mux.HandleFunc(models.QueryMetricValue, ds.HandleGetMetricValueQuery)
	mux.HandleFunc(models.QueryMetricHistory, ds.HandleGetMetricHistoryQuery)
	mux.HandleFunc(models.QueryMetricAggregate, ds.HandleGetMetricAggregate)

	ds.queryMux = mux
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
func (ds *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return ds.queryMux.QueryData(ctx, req)
}

func (ds *Datasource) SubscribeStream(ctx context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	logger := &GrafanaLogger{}
	logger.Info("SubscribeStream started", "path", req.Path, "data", string(req.Data))

	// Parse and validate the query
	parser := NewStreamQueryParser(ds.backendAPI)
	query, err := parser.ParseStreamQuery(req.Data)
	if err != nil {
		logger.Error("Failed to parse stream query in SubscribeStream", "error", err)
		return &backend.SubscribeStreamResponse{
			Status: backend.SubscribeStreamStatusNotFound,
		}, err
	}

	// Validate query and get streaming configuration from backend
	streamingConfig, err := parser.ValidateQuery(ctx, query)
	if err != nil {
		logger.Error("Invalid stream query in SubscribeStream", "error", err)
		return &backend.SubscribeStreamResponse{
			Status: backend.SubscribeStreamStatusNotFound,
		}, err
	}

	logger.Info("Validated stream query with backend configuration",
		"queryType", query.QueryType,
		"metrics", query.Metrics,
		"backendLookBackMs", streamingConfig.LookBackPeriodLimit,
		"backendLoopIntervalMs", streamingConfig.LoopInterval)

	// Create query executor
	factory := NewQueryExecutorFactory(ds.backendAPI)
	executor, err := factory.CreateExecutor(query)
	if err != nil {
		logger.Error("Failed to create query executor in SubscribeStream", "error", err)
		return &backend.SubscribeStreamResponse{
			Status: backend.SubscribeStreamStatusNotFound,
		}, err
	}

	// Get initial data using backend configuration
	initialFrames, err := ds.getInitialStreamDataWithConfig(ctx, executor, query, streamingConfig)
	if err != nil {
		logger.Error("Failed to get initial stream data", "error", err)
		return &backend.SubscribeStreamResponse{
			Status: backend.SubscribeStreamStatusNotFound,
		}, err
	}

	logger.Info("Successfully retrieved initial data", "frameCount", len(initialFrames))

	// Convert frames to initial data
	var initialData *backend.InitialData
	if len(initialFrames) > 0 {
		// Use the first frame as initial data
		var err error
		initialData, err = backend.NewInitialFrame(initialFrames[0], data.IncludeAll)
		if err != nil {
			logger.Error("Failed to create initial frame", "error", err)
			return &backend.SubscribeStreamResponse{
				Status: backend.SubscribeStreamStatusNotFound,
			}, err
		}

		// If there are multiple frames, we could potentially combine them
		// For now, we'll just use the first one
		if len(initialFrames) > 1 {
			logger.Info("Multiple frames received, using first frame for initial data", "totalFrames", len(initialFrames))
		}
	}

	return &backend.SubscribeStreamResponse{
		Status:      backend.SubscribeStreamStatusOK,
		InitialData: initialData,
	}, nil
}

func (ds *Datasource) PublishStream(context.Context, *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

type streamingConfig struct {
	LookBackPeriod *string `json:"lookBackPeriod,omitempty"`
}

type Q struct {
	QueryType string `json:"queryType"`
	// Range           backend.TimeRange
	IntervalMS      int64           `json:"intervalMs"`
	MaxDataPoints   int64           `json:"maxDataPoints"`
	StreamingConfig streamingConfig `json:"streamingConfig"`
	models.MetricBaseQuery
}

const defaultLookBack = time.Hour

func parseLookBackPeriod(raw *string, logger StreamLogger) time.Duration {
	if raw == nil {
		logger.Info("Using default lookback period for initial data", "duration", defaultLookBack.String())
		return defaultLookBack
	}
	if d, err := time.ParseDuration(*raw); err == nil {
		logger.Info("Using LookBackPeriod from streamingConfig for initial data", "lookBackPeriod", raw, "duration", d.String())
		return d
	}
	logger.Error("invalid lookback period specified -> using default lookback period", "lookBackPeriod", *raw)
	return defaultLookBack
}

// getInitialStreamDataWithConfig retrieves the initial dataset using backend configuration
func (ds *Datasource) getInitialStreamDataWithConfig(ctx context.Context, executor QueryExecutor, query *Q, backendConfig *models.StreamingQueryConfigurationResponse) (data.Frames, error) {
	logger := &GrafanaLogger{}
	now := time.Now()

	// Create stream config to resolve lookback period with backend limits
	streamConfig := NewStreamConfigFromBackend(backendConfig, query)

	timeRange := backend.TimeRange{
		From: now.Add(-streamConfig.LookBackPeriod),
		To:   now,
	}

	logger.Info("getInitialStreamDataWithConfig: Using resolved lookback period",
		"duration", streamConfig.LookBackPeriod.String(),
		"from", timeRange.From.Format(time.RFC3339),
		"to", timeRange.To.Format(time.RFC3339))

	return executor.ExecuteQuery(ctx, timeRange)
}

func (ds *Datasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	logger := &GrafanaLogger{}
	logger.Info("RunStream started", "req.Data", string(req.Data))

	// Parse and validate the query (this should have been done in SubscribeStream, but we validate again for safety)
	parser := NewStreamQueryParser(ds.backendAPI)
	query, err := parser.ParseStreamQuery(req.Data)
	if err != nil {
		logger.Error("Failed to parse stream query in RunStream", "error", err)
		return err
	}

	// Validate query and get streaming configuration from backend
	streamingConfig, err := parser.ValidateQuery(ctx, query)
	if err != nil {
		logger.Error("Invalid stream query in RunStream", "error", err)
		return err
	}

	logger.Info("Validated stream query for streaming with backend configuration",
		"queryType", query.QueryType,
		"metrics", query.Metrics,
		"backendLookBackMs", streamingConfig.LookBackPeriodLimit,
		"backendLoopIntervalMs", streamingConfig.LoopInterval)

	// Create query executor
	factory := NewQueryExecutorFactory(ds.backendAPI)
	executor, err := factory.CreateExecutor(query)
	if err != nil {
		logger.Error("Failed to create query executor in RunStream", "error", err)
		return err
	}

	// Create stream processor with backend configuration
	// Note: Initial data is already sent via SubscribeStream, so we skip that step
	processor := NewStreamProcessorFromBackendConfig(streamingConfig, query, executor, sender, logger)

	// Start the streaming loop directly (skip initial data since it was sent in SubscribeStream)
	logger.Info("Starting streaming loop with backend configuration (initial data already sent via SubscribeStream)",
		"tickInterval", processor.config.TickInterval.String())
	return processor.RunStreamingLoop(ctx)
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (ds *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	_, err := ds.backendAPI.GetDimensionKeys(ctx, models.GetDimensionKeysRequest{})
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

func (ds *Datasource) Dispose() {
	ds.backendAPI.Dispose()
}
