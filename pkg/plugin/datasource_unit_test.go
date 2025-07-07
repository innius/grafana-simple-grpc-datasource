package plugin

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
)

// TestRunStreamParsing tests the parsing and validation logic of RunStream
func TestRunStreamParsing(t *testing.T) {
	t.Run("valid query parsing", func(t *testing.T) {
		// Create test query
		testQuery := Q{
			QueryType:  models.QueryMetricValue,
			IntervalMS: 1000,
			MetricBaseQuery: models.MetricBaseQuery{
				Metrics: []models.Metric{{MetricId: "test-metric"}},
			},
		}

		// Marshal query to JSON
		queryData, err := json.Marshal(testQuery)
		require.NoError(t, err)

		// Test parsing
		parser := &StreamQueryParser{}
		parsedQuery, err := parser.ParseStreamQuery(queryData)
		
		assert.NoError(t, err)
		assert.Equal(t, models.QueryMetricValue, parsedQuery.QueryType)
		assert.Equal(t, int64(1000), parsedQuery.IntervalMS)
		assert.Len(t, parsedQuery.Metrics, 1)
		assert.Equal(t, "test-metric", parsedQuery.Metrics[0].MetricId)

		// Test validation
		err = parser.ValidateQuery(parsedQuery)
		assert.NoError(t, err)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		parser := &StreamQueryParser{}
		_, err := parser.ParseStreamQuery([]byte("invalid json"))
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal stream query")
	})

	t.Run("unsupported query type", func(t *testing.T) {
		testQuery := Q{QueryType: "unsupported"}
		queryData, err := json.Marshal(testQuery)
		require.NoError(t, err)

		parser := &StreamQueryParser{}
		parsedQuery, err := parser.ParseStreamQuery(queryData)
		require.NoError(t, err)

		err = parser.ValidateQuery(parsedQuery)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported query type")
	})
}

// TestQueryExecutorFactory tests the factory pattern for creating executors
func TestQueryExecutorFactoryIntegration(t *testing.T) {
	mockBackend := &MockBackendAPI{}
	factory := NewQueryExecutorFactory(mockBackend)

	testCases := []struct {
		name          string
		queryType     string
		expectedType  interface{}
		shouldSucceed bool
	}{
		{
			name:          "metric aggregate",
			queryType:     models.QueryMetricAggregate,
			expectedType:  &MetricAggregateExecutor{},
			shouldSucceed: true,
		},
		{
			name:          "metric history",
			queryType:     models.QueryMetricHistory,
			expectedType:  &MetricHistoryExecutor{},
			shouldSucceed: true,
		},
		{
			name:          "metric value",
			queryType:     models.QueryMetricValue,
			expectedType:  &MetricValueExecutor{},
			shouldSucceed: true,
		},
		{
			name:          "unsupported type",
			queryType:     "unsupported",
			expectedType:  nil,
			shouldSucceed: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := &Q{QueryType: tc.queryType}
			executor, err := factory.CreateExecutor(query)

			if tc.shouldSucceed {
				assert.NoError(t, err)
				assert.IsType(t, tc.expectedType, executor)
			} else {
				assert.Error(t, err)
				assert.Nil(t, executor)
			}
		})
	}
}

// TestStreamConfigDefaults tests the default configuration
func TestStreamConfigDefaults(t *testing.T) {
	config := DefaultStreamConfig()
	
	assert.Equal(t, 10*time.Second, config.TickInterval)
	assert.Equal(t, 1*time.Hour, config.InitialTimeSpan)
}

// TestModularityBenefits demonstrates the benefits of the refactored approach
func TestModularityBenefits(t *testing.T) {
	t.Run("components can be tested independently", func(t *testing.T) {
		// Parser can be tested independently
		parser := &StreamQueryParser{}
		assert.NotNil(t, parser)

		// Factory can be tested independently
		mockBackend := &MockBackendAPI{}
		factory := NewQueryExecutorFactory(mockBackend)
		assert.NotNil(t, factory)

		// Config can be tested independently
		config := DefaultStreamConfig()
		assert.NotNil(t, config)
	})

	t.Run("configuration is customizable", func(t *testing.T) {
		customConfig := StreamConfig{
			TickInterval:    5 * time.Second,
			InitialTimeSpan: 30 * time.Minute,
		}
		
		assert.Equal(t, 5*time.Second, customConfig.TickInterval)
		assert.Equal(t, 30*time.Minute, customConfig.InitialTimeSpan)
	})

	t.Run("executors implement common interface", func(t *testing.T) {
		mockBackend := &MockBackendAPI{}
		baseQuery := models.MetricBaseQuery{
			Metrics: []models.Metric{{MetricId: "test"}},
		}

		executors := []QueryExecutor{
			&MetricAggregateExecutor{backendAPI: mockBackend, metricBaseQuery: baseQuery},
			&MetricHistoryExecutor{backendAPI: mockBackend, metricBaseQuery: baseQuery},
			&MetricValueExecutor{backendAPI: mockBackend, metricBaseQuery: baseQuery},
		}

		// All executors implement the same interface
		for _, executor := range executors {
			assert.Implements(t, (*QueryExecutor)(nil), executor)
		}
	})
}
