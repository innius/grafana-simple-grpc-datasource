# Dynamic Interval Refactor for Stream Handler

## Overview

This refactor adds support for dynamic polling intervals in the streaming functionality, specifically designed for `MetricAggregateExecutor`. The key insight is that aggregate data often has natural intervals that should determine the polling frequency, rather than using a fixed interval.

## Problem Statement

Previously, all streaming queries used a fixed `TickInterval` from `StreamConfig`. This worked well for:
- `MetricHistoryExecutor` - Historical data with predictable intervals
- `MetricValueExecutor` - Current values that can be polled at regular intervals

However, `MetricAggregateExecutor` is different because:
- The backend returns an initial dataset with natural intervals
- The interval between the last two datapoints indicates the optimal polling frequency
- Using a fixed interval could result in over-polling or under-polling

## Solution

### New Interfaces and Types

1. **`DynamicIntervalExecutor` Interface**
   ```go
   type DynamicIntervalExecutor interface {
       QueryExecutor
       CalculateIntervalFromFrames(frames data.Frames) (time.Duration, error)
       SupportsDynamicInterval() bool
   }
   ```

2. **`DynamicStreamConfig` Struct**
   ```go
   type DynamicStreamConfig struct {
       InitialTimeSpan time.Duration // How much historical data to fetch initially
       MinInterval     time.Duration // Minimum allowed polling interval
       MaxInterval     time.Duration // Maximum allowed polling interval
   }
   ```

### Updated Executors

- **`MetricAggregateExecutor`**: Now implements `DynamicIntervalExecutor`
  - `SupportsDynamicInterval()` returns `true`
  - `CalculateIntervalFromFrames()` analyzes time series data to determine interval

- **`MetricHistoryExecutor`** and **`MetricValueExecutor`**: Implement the interface but don't support dynamic intervals
  - `SupportsDynamicInterval()` returns `false`
  - `CalculateIntervalFromFrames()` returns an error

### Enhanced StreamProcessor

The `StreamProcessor` now supports both fixed and dynamic intervals:

1. **Fixed Interval Mode** (existing behavior)
   - Created with `NewStreamProcessor()`
   - Uses `StreamConfig.TickInterval` for all polling

2. **Dynamic Interval Mode** (new behavior)
   - Created with `NewDynamicStreamProcessor()`
   - Calculates interval from initial data
   - Applies bounds checking (min/max limits)
   - Falls back to minimum interval if calculation fails

## Usage Examples

### Fixed Interval (Existing Behavior)
```go
config := StreamConfig{
    TickInterval:    10 * time.Second,
    InitialTimeSpan: 1 * time.Hour,
}

executor := &MetricHistoryExecutor{...}
processor := NewStreamProcessor(config, executor, sender, logger)
```

### Dynamic Interval (New Behavior)
```go
config := DynamicStreamConfig{
    InitialTimeSpan: 2 * time.Hour,
    MinInterval:     30 * time.Second,
    MaxInterval:     15 * time.Minute,
}

executor := &MetricAggregateExecutor{...}
processor := NewDynamicStreamProcessor(config, executor, sender, logger)
```

## How Dynamic Interval Calculation Works

1. **Initial Data Fetch**: Get historical data covering `InitialTimeSpan`
2. **Interval Detection**: Find the time difference between the last two datapoints
3. **Bounds Checking**: Ensure the calculated interval is within `MinInterval` and `MaxInterval`
4. **Polling**: Use the calculated interval for subsequent polling

### Example Calculation
```
Initial data timestamps:
- 12:00:00 (value: 45.2)
- 12:05:00 (value: 52.1)  <- Last two points
- 12:10:00 (value: 48.7)  <- are 5 minutes apart

Calculated interval: 5 minutes
If MinInterval = 1 minute and MaxInterval = 10 minutes:
-> Use 5 minutes (within bounds)
```

## Benefits

1. **Optimal Polling Frequency**: Matches the natural data interval
2. **Reduced Backend Load**: Avoids unnecessary polling
3. **Better Data Freshness**: Ensures timely updates when new data is available
4. **Backward Compatibility**: Existing code continues to work unchanged
5. **Configurable Bounds**: Prevents extreme polling intervals

## Testing

Comprehensive tests cover:
- Dynamic interval calculation from time series data
- Bounds checking (min/max enforcement)
- Fallback behavior when calculation fails
- Integration with existing streaming logic
- Backward compatibility with fixed intervals

## Migration Guide

### For MetricAggregateExecutor Users
```go
// Old approach (still works)
config := DefaultStreamConfig()
processor := NewStreamProcessor(config, aggregateExecutor, sender, logger)

// New recommended approach
dynamicConfig := DefaultDynamicStreamConfig()
processor := NewDynamicStreamProcessor(dynamicConfig, aggregateExecutor, sender, logger)
```

### For Other Executors
No changes required - continue using the existing `NewStreamProcessor()` approach.

## Implementation Details

### Key Files Modified
- `stream_handler.go`: Core implementation
- `stream_handler_test.go`: Comprehensive test coverage
- `dynamic_interval_example.go`: Usage examples

### Backward Compatibility
- All existing APIs remain unchanged
- Default configurations provide sensible defaults
- Graceful fallback when dynamic calculation fails

This refactor provides a more intelligent streaming solution while maintaining full backward compatibility and adding comprehensive test coverage.
