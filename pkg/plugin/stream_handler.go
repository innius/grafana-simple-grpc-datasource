package plugin

import (
	"context"
	"encoding/json"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/pkg/errors"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
)

// StreamConfig holds configuration for streaming operations
type StreamConfig struct {
	TickInterval    time.Duration
	InitialTimeSpan time.Duration
}

// DefaultStreamConfig returns default streaming configuration
func DefaultStreamConfig() StreamConfig {
	return StreamConfig{
		TickInterval:    10 * time.Second,
		InitialTimeSpan: 1 * time.Hour,
	}
}

// DynamicStreamConfig holds configuration for streaming operations with dynamic intervals
type DynamicStreamConfig struct {
	InitialTimeSpan time.Duration
	MinInterval     time.Duration // Minimum allowed interval
	MaxInterval     time.Duration // Maximum allowed interval
}

// DefaultDynamicStreamConfig returns default dynamic streaming configuration
func DefaultDynamicStreamConfig() DynamicStreamConfig {
	return DynamicStreamConfig{
		InitialTimeSpan: 1 * time.Hour,
		MinInterval:     1 * time.Second,
		MaxInterval:     1 * time.Hour,
	}
}

// QueryExecutor defines the interface for executing queries
type QueryExecutor interface {
	ExecuteQuery(ctx context.Context, timeRange backend.TimeRange) (data.Frames, error)
}

// DynamicIntervalExecutor extends QueryExecutor with dynamic interval calculation
type DynamicIntervalExecutor interface {
	QueryExecutor
	CalculateIntervalFromFrames(frames data.Frames) (time.Duration, error)
	SupportsDynamicInterval() bool
}

// StreamQueryParser handles parsing and validation of stream queries
type StreamQueryParser struct{}

// ParseStreamQuery parses the raw query data into a structured query
func (p *StreamQueryParser) ParseStreamQuery(rawData []byte) (*Q, error) {
	query := &Q{}
	if err := json.Unmarshal(rawData, query); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal stream query")
	}
	return query, nil
}

// ValidateQuery validates the parsed query
func (p *StreamQueryParser) ValidateQuery(query *Q) error {
	if query.QueryType == "" {
		return errors.New("query type is required")
	}

	supportedTypes := []string{
		models.QueryMetricAggregate,
		models.QueryMetricHistory,
		models.QueryMetricValue,
	}

	for _, supportedType := range supportedTypes {
		if query.QueryType == supportedType {
			return nil
		}
	}

	return errors.Errorf("unsupported query type: %s", query.QueryType)
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

// SupportsDynamicInterval returns true for MetricAggregateExecutor
func (e *MetricAggregateExecutor) SupportsDynamicInterval() bool {
	return true
}

// CalculateIntervalFromFrames calculates the interval between the last two datapoints
func (e *MetricAggregateExecutor) CalculateIntervalFromFrames(frames data.Frames) (time.Duration, error) {
	if len(frames) == 0 {
		return 0, errors.New("no frames provided")
	}

	// Find the first frame with time data
	for _, frame := range frames {
		if frame.TimeSeriesSchema().Type == data.TimeSeriesTypeWide {
			// Find the time field (usually the first field in time series)
			var timeField *data.Field
			for _, field := range frame.Fields {
				if field.Type() == data.FieldTypeTime {
					timeField = field
					break
				}
			}

			if timeField == nil || timeField.Len() < 2 {
				continue
			}

			// Get the last two timestamps
			lastIdx := timeField.Len() - 1
			secondLastIdx := lastIdx - 1

			lastTime, ok1 := timeField.At(lastIdx).(time.Time)
			secondLastTime, ok2 := timeField.At(secondLastIdx).(time.Time)

			if !ok1 || !ok2 {
				continue
			}

			interval := lastTime.Sub(secondLastTime)
			if interval > 0 {
				return interval, nil
			}
		}
	}

	return 0, errors.New("unable to calculate interval from frames")
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

// SupportsDynamicInterval returns false for MetricHistoryExecutor
func (e *MetricHistoryExecutor) SupportsDynamicInterval() bool {
	return false
}

// CalculateIntervalFromFrames is not implemented for MetricHistoryExecutor
func (e *MetricHistoryExecutor) CalculateIntervalFromFrames(frames data.Frames) (time.Duration, error) {
	return 0, errors.New("dynamic interval not supported for MetricHistoryExecutor")
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

// SupportsDynamicInterval returns false for MetricValueExecutor
func (e *MetricValueExecutor) SupportsDynamicInterval() bool {
	return false
}

// CalculateIntervalFromFrames is not implemented for MetricValueExecutor
func (e *MetricValueExecutor) CalculateIntervalFromFrames(frames data.Frames) (time.Duration, error) {
	return 0, errors.New("dynamic interval not supported for MetricValueExecutor")
}

// FrameSender defines the interface for sending frames
type FrameSender interface {
	SendFrame(frame *data.Frame, include data.FrameInclude) error
}

// StreamProcessor handles the streaming logic
type StreamProcessor struct {
	config          StreamConfig
	dynamicConfig   *DynamicStreamConfig
	executor        QueryExecutor
	sender          FrameSender
	logger          StreamLogger
	dynamicInterval time.Duration // Calculated interval for dynamic executors
}

// StreamLogger defines the interface for logging stream operations
type StreamLogger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// GrafanaLogger wraps the Grafana backend logger
type GrafanaLogger struct{}

func (l *GrafanaLogger) Info(msg string, keysAndValues ...interface{}) {
	backend.Logger.Info(msg, keysAndValues...)
}

func (l *GrafanaLogger) Error(msg string, keysAndValues ...interface{}) {
	backend.Logger.Error(msg, keysAndValues...)
}

// NewStreamProcessor creates a new stream processor with fixed interval
func NewStreamProcessor(config StreamConfig, executor QueryExecutor, sender FrameSender, logger StreamLogger) *StreamProcessor {
	return &StreamProcessor{
		config:   config,
		executor: executor,
		sender:   sender,
		logger:   logger,
	}
}

// NewDynamicStreamProcessor creates a new stream processor with dynamic interval support
func NewDynamicStreamProcessor(dynamicConfig DynamicStreamConfig, executor QueryExecutor, sender FrameSender, logger StreamLogger) *StreamProcessor {
	return &StreamProcessor{
		dynamicConfig: &dynamicConfig,
		executor:      executor,
		sender:        sender,
		logger:        logger,
	}
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
	var initialTimeSpan time.Duration

	if p.dynamicConfig != nil {
		initialTimeSpan = p.dynamicConfig.InitialTimeSpan
	} else {
		initialTimeSpan = p.config.InitialTimeSpan
	}

	timeRange := backend.TimeRange{
		From: now.Add(-initialTimeSpan),
		To:   now,
	}

	frames, err := p.executor.ExecuteQuery(ctx, timeRange)
	if err != nil {
		p.logger.Error("Initial query failure", "error", err)
		return err
	}

	// For dynamic interval executors, calculate the interval from initial data
	if dynamicExecutor, ok := p.executor.(DynamicIntervalExecutor); ok && dynamicExecutor.SupportsDynamicInterval() {
		if interval, err := dynamicExecutor.CalculateIntervalFromFrames(frames); err == nil {
			// Apply bounds checking
			if p.dynamicConfig != nil {
				if interval < p.dynamicConfig.MinInterval {
					interval = p.dynamicConfig.MinInterval
					p.logger.Info("Calculated interval below minimum, using minimum", "calculated", interval, "minimum", p.dynamicConfig.MinInterval)
				} else if interval > p.dynamicConfig.MaxInterval {
					interval = p.dynamicConfig.MaxInterval
					p.logger.Info("Calculated interval above maximum, using maximum", "calculated", interval, "maximum", p.dynamicConfig.MaxInterval)
				}
			}
			p.dynamicInterval = interval
			p.logger.Info("Dynamic interval calculated from initial data", "interval", interval)
		} else {
			p.logger.Error("Failed to calculate dynamic interval, using default", "error", err)
			// Fall back to a default interval if calculation fails
			if p.dynamicConfig != nil {
				p.dynamicInterval = p.dynamicConfig.MinInterval
			} else {
				p.dynamicInterval = 10 * time.Second
			}
		}
	}

	return p.sendFrames(frames, data.IncludeAll)
}

// runStreamingLoop runs the main streaming loop
func (p *StreamProcessor) runStreamingLoop(ctx context.Context) error {
	var tickInterval time.Duration

	// Determine which interval to use
	if dynamicExecutor, ok := p.executor.(DynamicIntervalExecutor); ok && dynamicExecutor.SupportsDynamicInterval() && p.dynamicInterval > 0 {
		tickInterval = p.dynamicInterval
		p.logger.Info("Using dynamic interval for streaming", "interval", tickInterval.String())
	} else {
		tickInterval = p.config.TickInterval
		p.logger.Info("Using fixed interval for streaming", "interval", tickInterval.String())
	}

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	lastTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Stream context canceled")
			return ctx.Err()
		case tickTime := <-ticker.C:
			if err := p.processStreamTick(ctx, lastTime, tickTime); err != nil {
				p.logger.Error("Stream tick processing failed", "error", err)
				// Continue processing despite errors
			}
			lastTime = tickTime
		}
	}
}

// processStreamTick processes a single stream tick
func (p *StreamProcessor) processStreamTick(ctx context.Context, fromTime, toTime time.Time) error {
	timeRange := backend.TimeRange{
		From: fromTime,
		To:   toTime,
	}

	frames, err := p.executor.ExecuteQuery(ctx, timeRange)
	if err != nil {
		return errors.Wrap(err, "query execution failed")
	}

	return p.sendFrames(frames, data.IncludeDataOnly)
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
