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

// QueryExecutor defines the interface for executing queries
type QueryExecutor interface {
	ExecuteQuery(ctx context.Context, timeRange backend.TimeRange) (data.Frames, error)
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

// NewStreamProcessor creates a new stream processor
func NewStreamProcessor(config StreamConfig, executor QueryExecutor, sender FrameSender, logger StreamLogger) *StreamProcessor {
	return &StreamProcessor{
		config:   config,
		executor: executor,
		sender:   sender,
		logger:   logger,
	}
}

// ProcessStream handles the streaming process
func (p *StreamProcessor) ProcessStream(ctx context.Context) error {
	// Send initial data
	if err := p.sendInitialData(ctx); err != nil {
		return errors.Wrap(err, "failed to send initial data")
	}

	// Start streaming loop
	return p.runStreamingLoop(ctx)
}

// sendInitialData sends the initial data frame
func (p *StreamProcessor) sendInitialData(ctx context.Context) error {
	now := time.Now()
	timeRange := backend.TimeRange{
		From: now.Add(-p.config.InitialTimeSpan),
		To:   now,
	}

	frames, err := p.executor.ExecuteQuery(ctx, timeRange)
	if err != nil {
		p.logger.Error("Initial query failure", "error", err)
		return err
	}

	return p.sendFrames(frames)
}

// runStreamingLoop runs the main streaming loop
func (p *StreamProcessor) runStreamingLoop(ctx context.Context) error {
	ticker := time.NewTicker(p.config.TickInterval)
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

	return p.sendFrames(frames)
}

// sendFrames sends multiple frames to the client
func (p *StreamProcessor) sendFrames(frames data.Frames) error {
	for _, frame := range frames {
		p.logger.Info("Sending frame", "frame.Name", frame.Name)
		if err := p.sender.SendFrame(frame, data.IncludeAll); err != nil {
			p.logger.Error("Failed to send frame", "error", err)
			return err
		}
	}
	return nil
}
