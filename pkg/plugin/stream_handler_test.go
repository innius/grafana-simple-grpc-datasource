package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
)

// Mock implementations for testing

type MockBackendAPI struct {
	mock.Mock
}

func (m *MockBackendAPI) GetStreamingQueryConfiguration(ctx context.Context, query *models.StreamingQueryConfigurationRequest) (*models.StreamingQueryConfigurationResponse, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*models.StreamingQueryConfigurationResponse), args.Error(1)
}

func (m *MockBackendAPI) HandleGetMetricAggregateQuery(ctx context.Context, query *models.MetricAggregateQuery) (data.Frames, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(data.Frames), args.Error(1)
}

func (m *MockBackendAPI) HandleGetMetricHistoryQuery(ctx context.Context, query *models.MetricHistoryQuery) (data.Frames, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(data.Frames), args.Error(1)
}

func (m *MockBackendAPI) HandleGetMetricValueQuery(ctx context.Context, query *models.MetricValueQuery) (data.Frames, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(data.Frames), args.Error(1)
}

func (m *MockBackendAPI) GetDimensionKeys(ctx context.Context, query models.GetDimensionKeysRequest) (*models.GetDimensionKeysResponse, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*models.GetDimensionKeysResponse), args.Error(1)
}

func (m *MockBackendAPI) GetDimensionValues(ctx context.Context, query models.GetDimensionValuesRequest) (*models.GetDimensionValueResponse, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*models.GetDimensionValueResponse), args.Error(1)
}

func (m *MockBackendAPI) GetMetrics(ctx context.Context, query models.GetMetricsRequest) (*models.GetMetricsResponse, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*models.GetMetricsResponse), args.Error(1)
}

func (m *MockBackendAPI) GetQueryOptions(ctx context.Context, input models.GetQueryOptionsRequest) (*models.GetQueryOptionsResponse, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.GetQueryOptionsResponse), args.Error(1)
}

func (m *MockBackendAPI) Dispose() {
	m.Called()
}

type MockFrameSender struct {
	mock.Mock
	SentFrames []*data.Frame
}

func (m *MockFrameSender) SendFrame(frame *data.Frame, include data.FrameInclude) error {
	m.SentFrames = append(m.SentFrames, frame)
	args := m.Called(frame, include)
	return args.Error(0)
}

type MockStreamLogger struct {
	mock.Mock
}

type LogEntry struct {
	Message       string
	KeysAndValues []interface{}
}

func (m *MockStreamLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockStreamLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockStreamLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

type MockQueryExecutor struct {
	mock.Mock
}

func (m *MockQueryExecutor) ExecuteQuery(ctx context.Context, timeRange models.TimeRange) (data.Frames, error) {
	args := m.Called(ctx, timeRange)
	return args.Get(0).(data.Frames), args.Error(1)
}

type MockDynamicIntervalExecutor struct {
	MockQueryExecutor
}

func (m *MockDynamicIntervalExecutor) SupportsDynamicInterval() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockDynamicIntervalExecutor) CalculateIntervalFromFrames(frames data.Frames) (time.Duration, error) {
	args := m.Called(frames)
	return args.Get(0).(time.Duration), args.Error(1)
}

// Test StreamQueryParser

func TestStreamQueryParser_ParseStreamQuery(t *testing.T) {
	parser := &StreamQueryParser{}

	t.Run("valid query", func(t *testing.T) {
		queryData := Q{
			QueryType: models.QueryMetricValue,
			// Range:      models.TimeRange{From: time.Now(), To: time.Now()},
			IntervalMS: 1000,
			MetricBaseQuery: models.MetricBaseQuery{
				Metrics: []models.Metric{{MetricId: "test-metric"}},
			},
		}

		rawData, err := json.Marshal(queryData)
		require.NoError(t, err)

		result, err := parser.ParseStreamQuery(rawData)

		assert.NoError(t, err)
		assert.Equal(t, models.QueryMetricValue, result.QueryType)
		assert.Equal(t, "test-metric", result.Metrics[0].MetricId)
		assert.Equal(t, int64(1000), result.IntervalMS)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		rawData := []byte("invalid json")

		result, err := parser.ParseStreamQuery(rawData)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to unmarshal stream query")
	})
}

func TestStreamQueryParser_ValidateQuery(t *testing.T) {
	parser := &StreamQueryParser{}

	t.Run("valid query types", func(t *testing.T) {
		validTypes := []string{
			models.QueryMetricAggregate,
			models.QueryMetricHistory,
			models.QueryMetricValue,
		}

		for _, queryType := range validTypes {
			query := &Q{QueryType: queryType}
			_, err := parser.ValidateQuery(context.TODO(), query)
			assert.NoError(t, err, "Query type %s should be valid", queryType)
		}
	})

	t.Run("empty query type", func(t *testing.T) {
		query := &Q{QueryType: ""}
		_, err := parser.ValidateQuery(context.TODO(), query)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query type is required")
	})

	t.Run("unsupported query type", func(t *testing.T) {
		query := &Q{QueryType: "unsupported"}
		_, err := parser.ValidateQuery(context.TODO(), query)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported query type: unsupported")
	})
}

// Test QueryExecutorFactory

func TestQueryExecutorFactory_CreateExecutor(t *testing.T) {
	mockBackend := &MockBackendAPI{}
	factory := NewQueryExecutorFactory(mockBackend)

	t.Run("metric aggregate executor", func(t *testing.T) {
		query := &Q{QueryType: models.QueryMetricAggregate}

		executor, err := factory.CreateExecutor(query)

		assert.NoError(t, err)
		assert.IsType(t, &MetricAggregateExecutor{}, executor)
	})

	t.Run("metric history executor", func(t *testing.T) {
		query := &Q{QueryType: models.QueryMetricHistory}

		executor, err := factory.CreateExecutor(query)

		assert.NoError(t, err)
		assert.IsType(t, &MetricHistoryExecutor{}, executor)
	})

	t.Run("metric value executor", func(t *testing.T) {
		query := &Q{QueryType: models.QueryMetricValue}

		executor, err := factory.CreateExecutor(query)

		assert.NoError(t, err)
		assert.IsType(t, &MetricValueExecutor{}, executor)
	})

	t.Run("unsupported query type", func(t *testing.T) {
		query := &Q{QueryType: "unsupported"}

		executor, err := factory.CreateExecutor(query)

		assert.Error(t, err)
		assert.Nil(t, executor)
		assert.Contains(t, err.Error(), "unsupported query type: unsupported")
	})
}

// Test Query Executors

func TestMetricAggregateExecutor_ExecuteQuery(t *testing.T) {
	mockBackend := &MockBackendAPI{}
	executor := &MetricAggregateExecutor{
		backendAPI: mockBackend,
		metricBaseQuery: models.MetricBaseQuery{
			Metrics: []models.Metric{{MetricId: "test-metric"}},
		},
	}

	ctx := context.Background()
	timeRange := models.TimeRange{From: time.Now(), To: time.Now()}
	expectedFrames := data.Frames{data.NewFrame("test")}

	mockBackend.On("HandleGetMetricAggregateQuery", ctx, mock.AnythingOfType("*models.MetricAggregateQuery")).Return(expectedFrames, nil)

	frames, err := executor.ExecuteQuery(ctx, timeRange)

	assert.NoError(t, err)
	assert.Equal(t, expectedFrames, frames)
	mockBackend.AssertExpectations(t)
}

func TestMetricHistoryExecutor_ExecuteQuery(t *testing.T) {
	mockBackend := &MockBackendAPI{}
	executor := &MetricHistoryExecutor{
		backendAPI: mockBackend,
		metricBaseQuery: models.MetricBaseQuery{
			Metrics: []models.Metric{{MetricId: "test-metric"}},
		},
	}

	ctx := context.Background()
	timeRange := models.TimeRange{From: time.Now(), To: time.Now()}
	expectedFrames := data.Frames{data.NewFrame("test")}

	mockBackend.On("HandleGetMetricHistoryQuery", ctx, mock.AnythingOfType("*models.MetricHistoryQuery")).Return(expectedFrames, nil)

	frames, err := executor.ExecuteQuery(ctx, timeRange)

	assert.NoError(t, err)
	assert.Equal(t, expectedFrames, frames)
	mockBackend.AssertExpectations(t)
}

func TestMetricValueExecutor_ExecuteQuery(t *testing.T) {
	mockBackend := &MockBackendAPI{}
	executor := &MetricValueExecutor{
		backendAPI: mockBackend,
		metricBaseQuery: models.MetricBaseQuery{
			Metrics: []models.Metric{{MetricId: "test-metric"}},
		},
	}

	ctx := context.Background()
	timeRange := models.TimeRange{From: time.Now(), To: time.Now()}
	expectedFrames := data.Frames{data.NewFrame("test")}

	mockBackend.On("HandleGetMetricValueQuery", ctx, mock.AnythingOfType("*models.MetricValueQuery")).Return(expectedFrames, nil)

	frames, err := executor.ExecuteQuery(ctx, timeRange)

	assert.NoError(t, err)
	assert.Equal(t, expectedFrames, frames)
	mockBackend.AssertExpectations(t)
}

// Test StreamProcessor

func TestStreamProcessor_ProcessStream(t *testing.T) {
	t.Run("successful initial data send", func(t *testing.T) {
		mockExecutor := &MockQueryExecutor{}
		mockSender := &MockFrameSender{}
		mockLogger := &MockStreamLogger{}

		config := StreamConfig{
			TickInterval:   100 * time.Millisecond,
			LookBackPeriod: 1 * time.Hour, // Add the LookBackPeriod field
		}

		processor := NewStreamProcessor(config, mockExecutor, mockSender, mockLogger)

		expectedFrames := data.Frames{data.NewFrame("test")}
		mockExecutor.On("ExecuteQuery", mock.Anything, mock.AnythingOfType("models.TimeRange")).Return(expectedFrames, nil).Once()
		mockSender.On("SendFrame", mock.Anything, data.IncludeAll).Return(nil).Once()
		mockLogger.On("Info", "sendInitialData: Using LookBackPeriod from StreamConfig", mock.Anything).Once()
		mockLogger.On("Info", mock.Anything, mock.Anything).Maybe()
		mockLogger.On("Error", mock.Anything, mock.Anything).Maybe()

		// Create a context that will be canceled quickly to avoid infinite loop
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := processor.ProcessStream(ctx)

		// Should return context.DeadlineExceeded due to timeout
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)

		// Verify initial data was sent
		assert.Len(t, mockSender.SentFrames, 1)
		mockExecutor.AssertExpectations(t)
		mockSender.AssertExpectations(t)
	})

	t.Run("initial data send failure", func(t *testing.T) {
		mockExecutor := &MockQueryExecutor{}
		mockSender := &MockFrameSender{}
		mockLogger := &MockStreamLogger{}

		config := DefaultStreamConfig()
		processor := NewStreamProcessor(config, mockExecutor, mockSender, mockLogger)

		expectedError := errors.New("query failed")
		mockExecutor.On("ExecuteQuery", mock.Anything, mock.AnythingOfType("models.TimeRange")).Return(data.Frames{}, expectedError)
		mockLogger.On("Info", "sendInitialData: Using LookBackPeriod from StreamConfig", mock.Anything).Once()
		mockLogger.On("Error", "Initial query failure", mock.Anything).Once()

		ctx := context.Background()
		err := processor.ProcessStream(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send initial data")
		mockExecutor.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

func TestStreamProcessor_sendFrames(t *testing.T) {
	mockExecutor := &MockQueryExecutor{}
	mockSender := &MockFrameSender{}
	mockLogger := &MockStreamLogger{}

	processor := NewStreamProcessor(DefaultStreamConfig(), mockExecutor, mockSender, mockLogger)

	t.Run("successful frame sending", func(t *testing.T) {
		frames := data.Frames{
			data.NewFrame("frame1"),
			data.NewFrame("frame2"),
		}

		mockSender.On("SendFrame", frames[0], data.IncludeAll).Return(nil).Once()
		mockSender.On("SendFrame", frames[1], data.IncludeAll).Return(nil).Once()
		mockLogger.On("Info", "Sending frame", mock.Anything).Return().Twice()

		err := processor.sendFrames(frames, data.IncludeAll)

		assert.NoError(t, err)
		mockSender.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("frame sending failure", func(t *testing.T) {
		frames := data.Frames{data.NewFrame("frame1")}
		expectedError := errors.New("send failed")

		mockSender.On("SendFrame", frames[0], data.IncludeAll).Return(expectedError).Once()
		mockLogger.On("Info", "Sending frame", mock.Anything).Return().Once()
		mockLogger.On("Error", "Failed to send frame", mock.Anything).Once()

		err := processor.sendFrames(frames, data.IncludeAll)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockSender.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// Test DefaultStreamConfig

func TestDefaultStreamConfig(t *testing.T) {
	config := DefaultStreamConfig()

	assert.Equal(t, 10*time.Second, config.TickInterval)
	assert.Equal(t, 1*time.Hour, config.LookBackPeriod)
}

func TestNewStreamConfigFromQuery_LookBackPeriod(t *testing.T) {
	toPtr := func(s string) *string {
		return &s
	}
	tests := []struct {
		name                    string
		lookBackPeriod          *string
		expectedLookBackPeriod  time.Duration
		expectedInitialTimeSpan time.Duration
	}{
		{
			name:                   "With custom LookBackPeriod",
			lookBackPeriod:         toPtr("5m"),
			expectedLookBackPeriod: 5 * time.Minute,
		},
		{
			name:                   "With zero LookBackPeriod (use default)",
			lookBackPeriod:         nil,
			expectedLookBackPeriod: 24 * time.Hour, // Default ServerLookBackPeriod when no backend config
		},
		{
			name:                   "With large LookBackPeriod",
			lookBackPeriod:         toPtr("2h"),
			expectedLookBackPeriod: 2 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &Q{
				QueryType: models.QueryMetricHistory,
				StreamingConfig: streamingConfig{
					LookBackPeriod: tt.lookBackPeriod,
				},
			}

			config := NewStreamConfigFromQuery(query)

			assert.Equal(t, tt.expectedLookBackPeriod, config.LookBackPeriod, "LookBackPeriod should match expected value")
			assert.Equal(t, 10*time.Second, config.TickInterval, "TickInterval should remain default")
		})
	}
}
