package plugin

import (
	"context"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"

	"bitbucket.org/innius/grafana-simple-grpc-datasource/pkg/models"
)

// ExampleUsage demonstrates how to use the dynamic interval functionality
func ExampleUsage() {
	// Create a mock backend API (in real usage, this would be your actual backend)
	var backendAPI BackendAPI

	// Example 1: Using fixed interval (existing behavior)
	// This works for MetricHistoryExecutor and MetricValueExecutor
	fixedConfig := StreamConfig{
		TickInterval:    10 * time.Second,
		InitialTimeSpan: 1 * time.Hour,
	}

	historyExecutor := &MetricHistoryExecutor{
		backendAPI: backendAPI,
		metricBaseQuery: models.MetricBaseQuery{
			Metrics: []models.Metric{{MetricId: "temperature"}},
		},
	}

	// Create processor with fixed interval
	var sender FrameSender
	var logger StreamLogger
	fixedProcessor := NewStreamProcessor(fixedConfig, historyExecutor, sender, logger)

	// Example 2: Using dynamic interval (new behavior)
	// This is specifically designed for MetricAggregateExecutor
	dynamicConfig := DynamicStreamConfig{
		InitialTimeSpan: 2 * time.Hour,  // Get 2 hours of initial data
		MinInterval:     30 * time.Second, // Don't poll faster than every 30 seconds
		MaxInterval:     15 * time.Minute, // Don't poll slower than every 15 minutes
	}

	aggregateExecutor := &MetricAggregateExecutor{
		backendAPI: backendAPI,
		metricBaseQuery: models.MetricBaseQuery{
			Metrics: []models.Metric{{MetricId: "cpu_usage"}},
		},
	}

	// Create processor with dynamic interval
	dynamicProcessor := NewDynamicStreamProcessor(dynamicConfig, aggregateExecutor, sender, logger)

	// Usage in context
	ctx := context.Background()

	// Fixed interval streaming (traditional approach)
	go func() {
		if err := fixedProcessor.ProcessStream(ctx); err != nil {
			// Handle error
		}
	}()

	// Dynamic interval streaming (new approach)
	go func() {
		if err := dynamicProcessor.ProcessStream(ctx); err != nil {
			// Handle error
		}
	}()

	// The dynamic processor will:
	// 1. Fetch initial data covering the last 2 hours
	// 2. Calculate the interval between the last two datapoints
	// 3. Use that interval for subsequent polling (bounded by min/max)
	// 4. Continue polling at the calculated interval
}

// ExampleDynamicIntervalCalculation shows how the interval calculation works
func ExampleDynamicIntervalCalculation() {
	executor := &MetricAggregateExecutor{}

	// Create sample frame with time series data
	frame := data.NewFrame("cpu_metrics")
	
	// Time field with 5-minute intervals
	timeField := data.NewField("time", nil, []time.Time{
		time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2023, 1, 1, 12, 5, 0, 0, time.UTC),  // 5 minutes later
		time.Date(2023, 1, 1, 12, 10, 0, 0, time.UTC), // 5 minutes later
	})
	
	// Value field
	valueField := data.NewField("cpu_usage", nil, []float64{45.2, 52.1, 48.7})
	
	frame.Fields = append(frame.Fields, timeField, valueField)
	frames := data.Frames{frame}

	// Calculate interval - should return 5 minutes
	interval, err := executor.CalculateIntervalFromFrames(frames)
	if err != nil {
		// Handle error
		return
	}

	// interval will be 5 * time.Minute
	_ = interval
}

// ExampleBoundsChecking shows how min/max bounds are applied
func ExampleBoundsChecking() {
	dynamicConfig := DynamicStreamConfig{
		InitialTimeSpan: 1 * time.Hour,
		MinInterval:     1 * time.Minute,  // Minimum 1 minute
		MaxInterval:     10 * time.Minute, // Maximum 10 minutes
	}

	// If calculated interval is 30 seconds (below minimum):
	// -> Will use 1 minute instead
	
	// If calculated interval is 15 minutes (above maximum):
	// -> Will use 10 minutes instead
	
	// If calculated interval is 5 minutes (within bounds):
	// -> Will use 5 minutes as calculated

	_ = dynamicConfig
}
