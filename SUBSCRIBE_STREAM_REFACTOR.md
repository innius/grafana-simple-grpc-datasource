# SubscribeStream and RunStream Refactoring

## Overview

This document describes the refactoring of the streaming functionality to properly separate concerns between `SubscribeStream` and `RunStream` methods according to Grafana's streaming architecture.

## Changes Made

### 1. Enhanced SubscribeStream Method

The `SubscribeStream` method now performs the following responsibilities:

- **Query Validation**: Parses and validates the incoming stream query
- **Initial Data Retrieval**: Executes the query to get the initial dataset
- **Initial Data Response**: Returns the initial data as part of the subscription response

#### Key Features:
- Comprehensive error handling with appropriate status codes
- Logging for debugging and monitoring
- Support for multiple frame handling (uses first frame for initial data)
- Proper error propagation to client

### 2. Refactored RunStream Method

The `RunStream` method has been streamlined to focus on:

- **Query Re-validation**: Safety check (though validation should have been done in SubscribeStream)
- **Streaming Loop**: Continuous data streaming without initial data retrieval
- **Stream Processing**: Delegates to the existing StreamProcessor infrastructure

#### Key Changes:
- Removed initial data retrieval (now handled by SubscribeStream)
- Added direct call to `RunStreamingLoop()` to skip initial data step
- Maintained existing streaming logic and configuration

### 3. Stream Handler Enhancements

Added new method to `StreamProcessor`:

```go
// RunStreamingLoop runs the main streaming loop without sending initial data
// This is used when initial data has already been sent via SubscribeStream
func (p *StreamProcessor) RunStreamingLoop(ctx context.Context) error
```

This allows the streaming loop to be started independently of initial data retrieval.

## Architecture Benefits

### Separation of Concerns
- **SubscribeStream**: Handles subscription setup and initial data
- **RunStream**: Handles continuous streaming

### Improved Error Handling
- Validation errors are caught early in SubscribeStream
- Client receives immediate feedback on invalid queries
- Streaming errors are handled separately from subscription errors

### Better Resource Management
- Initial data is retrieved once during subscription
- Streaming loop focuses on incremental updates
- Reduced redundant data retrieval

## Usage Flow

1. **Client Subscribes**: Calls SubscribeStream with query
2. **Validation**: Query is parsed and validated
3. **Initial Data**: Historical data is retrieved and returned
4. **Stream Start**: RunStream begins continuous streaming
5. **Updates**: Only new/updated data is streamed

## Testing

Comprehensive integration tests have been added:

- `TestSubscribeStreamIntegration`: Tests successful subscription with initial data
- `TestSubscribeStreamValidationFailure`: Tests error handling for invalid queries
- `TestRunStreamQueryValidation`: Tests RunStream error handling

## Backward Compatibility

The refactoring maintains backward compatibility:
- Existing query structures are preserved
- Stream processing logic remains unchanged
- Configuration options are maintained

## Error Handling

### SubscribeStream Errors
- Returns `SubscribeStreamStatusNotFound` for validation failures
- Includes detailed error messages in logs
- Prevents invalid streams from starting

### RunStream Errors
- Continues existing error handling patterns
- Maintains streaming resilience
- Provides detailed logging for debugging

## Configuration

The refactoring uses existing configuration:
- `DefaultStreamConfig()` for fixed intervals
- `DefaultDynamicStreamConfig()` for dynamic intervals
- Existing timeout and retry logic

## Future Enhancements

This refactoring enables future improvements:
- Enhanced initial data customization
- Better subscription management
- Improved streaming performance monitoring
- Advanced error recovery mechanisms
