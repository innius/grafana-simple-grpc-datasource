package plugin

import (
	"context"
	"encoding/json"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/pkg/errors"

	"slices"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
)

// StreamConfig holds configuration for streaming operations
type StreamConfig struct {
	TickInterval      time.Duration
	LookBackPeriod    time.Duration
	MaxLookBackPeriod time.Duration // Maximum allowed by backend
}

// DefaultStreamConfig returns default streaming configuration
func DefaultStreamConfig() StreamConfig {
	return StreamConfig{
		TickInterval:      10 * time.Second, // Default TickInterval
		LookBackPeriod:    1 * time.Hour,    // Default lookback period
		MaxLookBackPeriod: 24 * time.Hour,   // Default maximum
	}
}

// NewStreamConfigFromBackend creates a StreamConfig from backend configuration and query preferences
func NewStreamConfigFromBackend(backendConfig *models.StreamingQueryConfigurationResponse, query *Q) StreamConfig {
	config := DefaultStreamConfig()

	// Use backend configuration if available
	if backendConfig != nil {
		if backendConfig.LoopInterval > 0 {
			config.TickInterval = backendConfig.LoopInterval
			backend.Logger.Info("StreamConfig: Using TickInterval from backend",
				"intervalMs", backendConfig.LoopInterval,
				"duration", config.TickInterval.String())
		}

		if backendConfig.LookBackPeriodLimit > 0 {
			config.MaxLookBackPeriod = backendConfig.LookBackPeriodLimit
			backend.Logger.Info("StreamConfig: Backend MaxLookBackPeriod",
				"maxLookBackMs", backendConfig.LookBackPeriodLimit,
				"duration", config.MaxLookBackPeriod.String())
		}
	}

	// Resolve client preference for lookback period
	if query.StreamingConfig.LookBackPeriod != nil {
		clientLookBack := parseLookBackPeriod(query.StreamingConfig.LookBackPeriod, backend.Logger)

		// Respect backend maximum if set
		if config.MaxLookBackPeriod > 0 && clientLookBack > config.MaxLookBackPeriod {
			config.LookBackPeriod = config.MaxLookBackPeriod
			backend.Logger.Info("StreamConfig: Client LookBackPeriod exceeds backend maximum, using maximum",
				"clientRequested", clientLookBack.String(),
				"backendMaximum", config.MaxLookBackPeriod.String(),
				"using", config.LookBackPeriod.String())
		} else {
			config.LookBackPeriod = clientLookBack
			backend.Logger.Info("StreamConfig: Using client LookBackPeriod",
				"lookBackPeriod", query.StreamingConfig.LookBackPeriod,
				"duration", config.LookBackPeriod.String())
		}
	} else {
		// Use default, but respect backend maximum
		if config.MaxLookBackPeriod > 0 && config.LookBackPeriod > config.MaxLookBackPeriod {
			config.LookBackPeriod = config.MaxLookBackPeriod
			backend.Logger.Info("StreamConfig: Default LookBackPeriod exceeds backend maximum, using maximum",
				"default", 1*time.Hour,
				"backendMaximum", config.MaxLookBackPeriod.String(),
				"using", config.LookBackPeriod.String())
		} else {
			backend.Logger.Info("StreamConfig: Using default LookBackPeriod", "duration", config.LookBackPeriod.String())
		}
	}

	return config
}

// NewStreamConfigFromQuery creates a StreamConfig from query streamingConfig (legacy method)
func NewStreamConfigFromQuery(query *Q) StreamConfig {
	return NewStreamConfigFromBackend(nil, query)
}

// QueryExecutor defines the interface for executing queries
type QueryExecutor interface {
	ExecuteQuery(ctx context.Context, timeRange backend.TimeRange) (data.Frames, error)
}

func NewStreamQueryParser(backendAPI BackendAPI) *StreamQueryParser {
	return &StreamQueryParser{
		backendAPI: backendAPI,
	}
}

// StreamQueryParser handles parsing and validation of stream queries
type StreamQueryParser struct {
	backendAPI BackendAPI
}

// ParseStreamQuery parses the raw query data into a structured query
func (p *StreamQueryParser) ParseStreamQuery(rawData []byte) (*Q, error) {
	query := &Q{}
	if err := json.Unmarshal(rawData, query); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal stream query")
	}
	return query, nil
}

var supportedQueryTypes = []string{
	models.QueryMetricAggregate,
	models.QueryMetricHistory,
	models.QueryMetricValue,
}

func defaultStreamingConfig() *models.StreamingQueryConfigurationResponse {
	return &models.StreamingQueryConfigurationResponse{
		LookBackPeriodLimit: time.Hour,
		LoopInterval:        10 * time.Second,
	}
}

func queryModelForType(queryType string, baseQuery models.MetricBaseQuery) interface{} {
	switch queryType {
	case models.QueryMetricValue:
		return models.MetricValueQuery{MetricBaseQuery: baseQuery}
	case models.QueryMetricHistory:
		return models.MetricHistoryQuery{MetricBaseQuery: baseQuery}
	case models.QueryMetricAggregate:
		return models.MetricAggregateQuery{MetricBaseQuery: baseQuery}
	default:
		return nil
	}
}

func (p *StreamQueryParser) ValidateQuery(ctx context.Context, query *Q) (*models.StreamingQueryConfigurationResponse, error) {
	if query.QueryType == "" {
		return nil, errors.New("query type is required")
	}

	if !slices.Contains(supportedQueryTypes, query.QueryType) {
		return nil, errors.Errorf("unsupported query type: %s", query.QueryType)
	}

	if p.backendAPI != nil {
		queryModel := queryModelForType(query.QueryType, query.MetricBaseQuery)
		configReq := &models.StreamingQueryConfigurationRequest{
			Query: queryModel,
		}

		cfg, err := p.backendAPI.GetStreamingQueryConfiguration(ctx, configReq)
		if err != nil {
			backend.Logger.Warn("Failed to get streaming configuration from backend, using defaults", "error", err)
			return defaultStreamingConfig(), nil
		}

		return cfg, nil
	}

	backend.Logger.Warn("Using default streaming configuration")
	return defaultStreamingConfig(), nil
}

// QueryExecutorFactory creates query executors based on query type
type QueryExecutorFactory struct {
	backendAPI BackendAPI
}

// BackendAPI defines the interface for backend operations
type BackendAPI interface {
	HandleGetMetricAggregateQuery(ctx context.Context, query *models.MetricAggregateQuery) (data.Frames, error)
	HandleGetMetricHistoryQuery(ctx context.Context, query *models.MetricHistoryQuery) (data.Frames, error)
	HandleGetMetricValueQuery(ctx context.Context, query *models.MetricValueQuery) (data.Frames, error)
	GetStreamingQueryConfiguration(ctx context.Context, query *models.StreamingQueryConfigurationRequest) (*models.StreamingQueryConfigurationResponse, error)
}

// NewQueryExecutorFactory creates a new query executor factory
func NewQueryExecutorFactory(backendAPI BackendAPI) *QueryExecutorFactory {
	return &QueryExecutorFactory{
		backendAPI: backendAPI,
	}
}

// CreateExecutor creates a query executor for the given query
func (f *QueryExecutorFactory) CreateExecutor(query *Q) (QueryExecutor, error) {
	switch query.QueryType {
	case models.QueryMetricAggregate:
		return &MetricAggregateExecutor{
			backendAPI:      f.backendAPI,
			metricBaseQuery: query.MetricBaseQuery,
		}, nil
	case models.QueryMetricHistory:
		return &MetricHistoryExecutor{
			backendAPI:      f.backendAPI,
			metricBaseQuery: query.MetricBaseQuery,
		}, nil
	case models.QueryMetricValue:
		return &MetricValueExecutor{
			backendAPI:      f.backendAPI,
			metricBaseQuery: query.MetricBaseQuery,
		}, nil
	default:
		return nil, errors.Errorf("unsupported query type: %s", query.QueryType)
	}
}

// MetricAggregateExecutor executes metric aggregate queries
type MetricAggregateExecutor struct {
	backendAPI      BackendAPI
	metricBaseQuery models.MetricBaseQuery
}

func (e *MetricAggregateExecutor) ExecuteQuery(ctx context.Context, timeRange backend.TimeRange) (data.Frames, error) {
	query := &models.MetricAggregateQuery{
		MetricBaseQuery: e.metricBaseQuery,
	}
	query.TimeRange = timeRange
	return e.backendAPI.HandleGetMetricAggregateQuery(ctx, query)
}

// MetricHistoryExecutor executes metric history queries
type MetricHistoryExecutor struct {
	backendAPI      BackendAPI
	metricBaseQuery models.MetricBaseQuery
}

func (e *MetricHistoryExecutor) ExecuteQuery(ctx context.Context, timeRange backend.TimeRange) (data.Frames, error) {
	query := &models.MetricHistoryQuery{
		MetricBaseQuery: e.metricBaseQuery,
	}
	query.TimeRange = timeRange
	return e.backendAPI.HandleGetMetricHistoryQuery(ctx, query)
}

// MetricValueExecutor executes metric value queries
type MetricValueExecutor struct {
	backendAPI      BackendAPI
	metricBaseQuery models.MetricBaseQuery
}

func (e *MetricValueExecutor) ExecuteQuery(ctx context.Context, timeRange backend.TimeRange) (data.Frames, error) {
	query := &models.MetricValueQuery{
		MetricBaseQuery: e.metricBaseQuery,
	}
	query.TimeRange = timeRange
	return e.backendAPI.HandleGetMetricValueQuery(ctx, query)
}

// FrameSender defines the interface for sending frames
type FrameSender interface {
	SendFrame(frame *data.Frame, include data.FrameInclude) error
}

// StreamProcessor handles the streaming logic
type StreamProcessor struct {
	config   StreamConfig
	executor QueryExecutor
	sender   FrameSender
	logger   StreamLogger
}

// StreamLogger defines the interface for logging stream operations
type StreamLogger interface {
	Info(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}

// // GrafanaLogger wraps the Grafana backend logger
// type GrafanaLogger struct{}
//
// func (l *GrafanaLogger) Info(msg string, keysAndValues ...any) {
// 	backend.Logger.Info(msg, keysAndValues...)
// }
//
// func (l *GrafanaLogger) Debug(msg string, keysAndValues ...any) {
// 	backend.Logger.Debug(msg, keysAndValues...)
// }
//
// func (l *GrafanaLogger) Error(msg string, keysAndValues ...any) {
// 	backend.Logger.Error(msg, keysAndValues...)
// }

// NewStreamProcessor creates a new stream processor with fixed interval
func NewStreamProcessor(config StreamConfig, executor QueryExecutor, sender FrameSender, logger StreamLogger) *StreamProcessor {
	return &StreamProcessor{
		config:   config,
		executor: executor,
		sender:   sender,
		logger:   logger,
	}
}

// NewStreamProcessorFromQuery creates a stream processor using configuration from the query (legacy)
func NewStreamProcessorFromQuery(query *Q, executor QueryExecutor, sender FrameSender, logger StreamLogger) *StreamProcessor {
	config := NewStreamConfigFromQuery(query)
	return NewStreamProcessor(config, executor, sender, logger)
}

// NewStreamProcessorFromBackendConfig creates a stream processor using backend configuration
func NewStreamProcessorFromBackendConfig(backendConfig *models.StreamingQueryConfigurationResponse, query *Q, executor QueryExecutor, sender FrameSender, logger StreamLogger) *StreamProcessor {
	config := NewStreamConfigFromBackend(backendConfig, query)
	return NewStreamProcessor(config, executor, sender, logger)
}

// ProcessStream handles the streaming process (including initial data)
func (p *StreamProcessor) ProcessStream(ctx context.Context) error {
	// Send initial data
	if err := p.sendInitialData(ctx); err != nil {
		return errors.Wrap(err, "failed to send initial data")
	}

	// Start streaming loop
	return p.RunStreamingLoop(ctx)
}

// RunStreamingLoop runs the main streaming loop without sending initial data
// This is used when initial data has already been sent via SubscribeStream
func (p *StreamProcessor) RunStreamingLoop(ctx context.Context) error {
	return p.runStreamingLoop(ctx)
}

// sendInitialData sends the initial data frame
func (p *StreamProcessor) sendInitialData(ctx context.Context) error {
	now := time.Now()
	lookBackPeriod := p.config.LookBackPeriod
	p.logger.Info("sendInitialData: Using LookBackPeriod from StreamConfig", "duration", lookBackPeriod.String())

	timeRange := backend.TimeRange{
		From: now.Add(-lookBackPeriod),
		To:   now,
	}

	frames, err := p.executor.ExecuteQuery(ctx, timeRange)
	if err != nil {
		p.logger.Error("Initial query failure", "error", err)
		return err
	}

	return p.sendFrames(frames, data.IncludeAll)
}

// runStreamingLoop runs the main streaming loop
func (p *StreamProcessor) runStreamingLoop(ctx context.Context) error {
	tickInterval := p.config.TickInterval
	p.logger.Info("Using fixed interval for streaming", "interval", tickInterval.String())

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	lastTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Stream context canceled")
			return ctx.Err()
		case tickTime := <-ticker.C:
			lastTime = p.handleTick(ctx, lastTime, tickTime)
		}
	}
}

func (p *StreamProcessor) handleTick(ctx context.Context, fromTime, toTime time.Time) time.Time {
	lastTime, err := p.processStreamTick(ctx, fromTime, toTime)
	if err != nil {
		p.logger.Error("Stream tick processing failed", "error", err)
		return fromTime
	}
	return lastTime
}

// processStreamTick processes a single stream tick
func (p *StreamProcessor) processStreamTick(ctx context.Context, fromTime, toTime time.Time) (time.Time, error) {
	timeRange := backend.TimeRange{
		From: fromTime,
		To:   toTime,
	}

	frames, err := p.executor.ExecuteQuery(ctx, timeRange)
	if err != nil {
		return toTime, errors.Wrap(err, "query execution failed")
	}
	if len(frames) == 0 {
		p.logger.Debug("No frames received from query execution", "time_range", timeRange)
		return fromTime, nil
	}
	return toTime, p.sendFrames(frames, data.IncludeDataOnly)
}

// sendFrames sends multiple frames to the client
func (p *StreamProcessor) sendFrames(frames data.Frames, include data.FrameInclude) error {
	for _, frame := range frames {
		p.logger.Info("Sending frame", "frame.Name", frame.Name)
		if err := p.sender.SendFrame(frame, include); err != nil {
			p.logger.Error("Failed to send frame", "error", err)
			return err
		}
	}
	return nil
}
