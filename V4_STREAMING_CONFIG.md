# V4 API Streaming Configuration

This document describes the enhanced streaming configuration feature introduced
with the v4 API for the Grafana Simple gRPC Datasource Plugin.

## Overview

The v4 API introduces a new `GetStreamingQueryConfiguration` operation that
allows backend systems to provide streaming configuration parameters, including:

- **Polling Interval**: How frequently the datasource should poll for new data
  during streaming (RunStream)
- **Maximum Look-back Period**: The maximum allowed initial historical data
  period (SubscribeStream)

This enables backend systems to control streaming behavior based on their
capabilities and constraints, while still allowing client preferences within
those limits.

## V4 API Changes

### New Proto Operation

```protobuf
service GrafanaQueryAPI {
  // ... existing operations
  
  rpc GetStreamingQueryConfiguration(GetStreamingQueryConfigurationRequest) 
    returns (GetStreamingQueryConfigurationResponse) {}
}

message GetStreamingQueryConfigurationRequest {
  oneof query {
    GetMetricValueRequest getMetricValueRequest = 1;
    GetMetricHistoryRequest getMetricHistoryRequest = 2;
    GetMetricAggregateRequest getMetricAggregateRequest = 3;
  }
}

message GetStreamingQueryConfigurationResponse {
  int64 lookBackPeriod = 1;  // Maximum lookback period in milliseconds
  int64 loopInterval = 2;    // Polling interval in milliseconds
}
```

### Backend Implementation Example

```go
func (s *MyBackendServer) GetStreamingQueryConfiguration(
    ctx context.Context, 
    req *v4.GetStreamingQueryConfigurationRequest,
) (*v4.GetStreamingQueryConfigurationResponse, error) {
    // Analyze the query to determine appropriate streaming configuration
    switch req.Query.(type) {
    case *v4.GetStreamingQueryConfigurationRequest_GetMetricValueRequest:
        floep := time.Minute
        // For metric value queries, allow shorter lookback and faster polling
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: 30*time.Millisecond
            LoopInterval:   5 * time.Second 
        }, nil
        
    case *v4.GetStreamingQueryConfigurationRequest_GetMetricHistoryRequest:
        // For history queries, allow longer lookback but slower polling
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: int64(24 * time.Hour / time.Millisecond),    // 24 hours max
            LoopInterval:   int64(30 * time.Second / time.Millisecond),  // Poll every 30 seconds
        }, nil
        
    case *v4.GetStreamingQueryConfigurationRequest_GetMetricAggregateRequest:
        // For aggregate queries, moderate settings
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: int64(6 * time.Hour / time.Millisecond),     // 6 hours max
            LoopInterval:   int64(15 * time.Second / time.Millisecond),  // Poll every 15 seconds
        }, nil
        
    default:
        return nil, status.Errorf(codes.InvalidArgument, "unsupported query type")
    }
}
```

## Configuration Resolution Logic

The datasource resolves streaming configuration using the following priority:

1. **Backend Limits**: The v4 API provides maximum constraints
2. **Client Preferences**: User-specified configuration within backend limits
3. **Defaults**: Fallback values when no configuration is provided

### Look-back Period Resolution

```
Final Look-back Period = min(Client Preference, Backend Maximum, Default Maximum)
```

**Examples:**

- Client wants "2h", Backend allows "1h" → Use "1h" (backend limit)
- Client wants "30m", Backend allows "2h" → Use "30m" (client preference)
- Client specifies nothing, Backend allows "6h" → Use "1h" (default, within
  limit)
- Client specifies nothing, Backend allows "30m" → Use "30m" (backend limit)

### Polling Interval Resolution

```
Final Polling Interval = Backend Specified OR Default (10 seconds)
```

The backend has full control over polling intervals to manage load and resource
usage.

## Frontend Integration

### Automatic Configuration Fetching

The frontend automatically fetches streaming configuration when a streaming
query is executed:

```typescript
async runGrafanaLiveQuery(target: MyQuery, req: DataQueryRequest<MyQuery>): Promise<Observable<DataQueryResponse>> {
  // Fetch backend streaming configuration
  const backendConfig = await this.getStreamingConfiguration(target);
  
  if (backendConfig) {
    // Merge with client preferences, respecting backend limits
    streamingConfig = {
      ...streamingConfig,
      maxLookBackPeriod: backendConfig.maxLookBackPeriod,
      backendPollingInterval: backendConfig.backendPollingInterval,
      lookBackPeriod: this.validateLookBackPeriod(
        streamingConfig.lookBackPeriod, 
        backendConfig.maxLookBackPeriod
      ),
    };
  }
  
  // ... continue with streaming setup
}
```

### UI Enhancements

The streaming configuration editor now shows:

1. **Backend Limits**: Information panel showing backend-provided constraints
2. **Validation**: Real-time validation of client preferences against backend
   limits
3. **Warnings**: Visual indicators when client preferences exceed backend limits

```typescript
// Backend configuration display
{
  hasBackendLimits && (
    <Alert title="Backend Configuration" severity="info">
      {config.maxLookBackPeriod && (
        <div>
          Maximum lookback period: {formatDuration(config.maxLookBackPeriod)}
        </div>
      )}
      {config.backendPollingInterval && (
        <div>
          Backend polling interval:{" "}
          {formatDuration(config.backendPollingInterval)}
        </div>
      )}
    </Alert>
  );
}
```

## Backend Implementation

### Client Detection

The datasource automatically detects v4 API support:

```go
// Try to create v4 client (optional)
v4Client, v4Err := v4client.NewClient(conn)
hasV4 := v4Err == nil

if hasV4 {
    log.DefaultLogger.Info("v4 API client available - streaming configuration supported")
} else {
    log.DefaultLogger.Info("v4 API client not available - using default streaming configuration")
}
```

### Graceful Fallback

When v4 API is not available, the datasource falls back to default
configuration:

```go
func GetStreamingQueryConfiguration(ctx context.Context, client client.BackendAPIClient, query models.StreamingQueryConfigurationRequest) (*models.StreamingQueryConfigurationResponse, error) {
    // Check if client supports v4 API
    v4Client, ok := client.GetV4Client()
    if !ok {
        // Fallback to default configuration if v4 is not supported
        return &models.StreamingQueryConfigurationResponse{
            LookBackPeriod: int64(3600000), // 1 hour in milliseconds
            LoopInterval:   int64(10000),   // 10 seconds in milliseconds
        }, nil
    }
    
    // ... use v4 client
}
```

### Stream Processing Integration

The stream processor uses backend configuration:

```go
// Create stream processor with backend configuration
processor := NewStreamProcessorFromBackendConfig(streamingConfig, query, executor, sender, logger)

// Configuration includes both polling interval and lookback limits
type StreamConfig struct {
    TickInterval      time.Duration // From backend LoopInterval
    LookBackPeriod    time.Duration // Resolved client preference
    MaxLookBackPeriod time.Duration // Backend maximum
}
```

## Migration and Compatibility

### Backwards Compatibility

- **V1/V2/V3 Backends**: Continue to work with default streaming configuration
- **Existing Queries**: No changes required, enhanced automatically when v4
  backend is available
- **Client Configuration**: All existing streaming configuration options remain
  functional

### Migration Path

1. **Backend Upgrade**: Implement `GetStreamingQueryConfiguration` in your
   backend
2. **No Client Changes**: Frontend automatically detects and uses v4
   capabilities
3. **Gradual Rollout**: Can deploy v4 backend without requiring client updates

### Version Detection

The datasource logs the API version being used:

```
INFO v4 API client available - streaming configuration supported
INFO v4 API client not available - using default streaming configuration
```

## Use Cases

### High-Frequency Monitoring

Backend can optimize for real-time monitoring:

```go
// For critical metrics, allow frequent polling but limit history
return &v4.GetStreamingQueryConfigurationResponse{
    LookBackPeriod: int64(15 * time.Minute / time.Millisecond), // 15 minutes
    LoopInterval:   int64(1 * time.Second / time.Millisecond),   // Every second
}
```

### Historical Analysis

Backend can optimize for historical data analysis:

```go
// For analytical queries, allow long history but slower updates
return &v4.GetStreamingQueryConfigurationResponse{
    LookBackPeriod: int64(7 * 24 * time.Hour / time.Millisecond), // 7 days
    LoopInterval:   int64(5 * time.Minute / time.Millisecond),     // Every 5 minutes
}
```

### Resource Management

Backend can manage load based on system capacity:

```go
func (s *Server) GetStreamingQueryConfiguration(ctx context.Context, req *v4.GetStreamingQueryConfigurationRequest) (*v4.GetStreamingQueryConfigurationResponse, error) {
    // Check current system load
    if s.isHighLoad() {
        // Reduce frequency and history during high load
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: int64(1 * time.Hour / time.Millisecond),
            LoopInterval:   int64(30 * time.Second / time.Millisecond),
        }, nil
    }
    
    // Normal load configuration
    return &v4.GetStreamingQueryConfigurationResponse{
        LookBackPeriod: int64(6 * time.Hour / time.Millisecond),
        LoopInterval:   int64(10 * time.Second / time.Millisecond),
    }, nil
}
```

## Testing

### Backend Testing

Test the streaming configuration endpoint:

```go
func TestGetStreamingQueryConfiguration(t *testing.T) {
    server := &MyBackendServer{}
    
    req := &v4.GetStreamingQueryConfigurationRequest{
        Query: &v4.GetStreamingQueryConfigurationRequest_GetMetricValueRequest{
            GetMetricValueRequest: &v4.GetMetricValueRequest{
                Metrics: []string{"cpu.usage"},
            },
        },
    }
    
    resp, err := server.GetStreamingQueryConfiguration(context.Background(), req)
    assert.NoError(t, err)
    assert.Greater(t, resp.LookBackPeriod, int64(0))
    assert.Greater(t, resp.LoopInterval, int64(0))
}
```

### Frontend Testing

Test configuration resolution:

```typescript
describe("Streaming Configuration", () => {
  it("should respect backend maximum lookback period", async () => {
    const datasource = new DataSource(instanceSettings);

    // Mock backend response
    jest.spyOn(datasource, "postResource").mockResolvedValue({
      lookBackPeriod: 3600000, // 1 hour
      loopInterval: 10000, // 10 seconds
    });

    const query: MyQuery = {
      streamingConfig: {
        lookBackPeriod: "2h", // Client wants 2 hours
      },
    };

    const config = await datasource.getStreamingConfiguration(query);

    // Should be limited to backend maximum
    expect(config?.maxLookBackPeriod).toBe("3600000ms");
  });
});
```

## Troubleshooting

### Common Issues

1. **V4 API Not Detected**
   - Check backend implements gRPC reflection
   - Verify v4 proto is properly compiled
   - Check connection logs for v4 client creation

2. **Configuration Not Applied**
   - Verify backend returns valid millisecond values
   - Check frontend logs for configuration fetching
   - Ensure streaming is enabled for the query

3. **Lookback Period Ignored**
   - Check if client preference exceeds backend maximum
   - Verify duration format (e.g., "1h", "30m", "5s")
   - Check for validation warnings in UI

### Debug Logging

Enable debug logging to troubleshoot:

```go
// Backend
log.DefaultLogger.Info("Streaming configuration requested", 
    "queryType", req.Query,
    "lookBackPeriod", resp.LookBackPeriod,
    "loopInterval", resp.LoopInterval)

// Frontend
console.log('Backend streaming config:', backendConfig);
console.log('Resolved streaming config:', streamingConfig);
```

## Best Practices

### Backend Implementation

1. **Query-Specific Configuration**: Tailor settings based on query type and
   complexity
2. **Resource Awareness**: Adjust limits based on current system load
3. **Reasonable Defaults**: Provide sensible fallback values
4. **Validation**: Validate configuration values before returning

### Frontend Usage

1. **Graceful Degradation**: Handle v4 API unavailability gracefully
2. **User Feedback**: Show backend limits and validation messages
3. **Performance**: Cache configuration when appropriate
4. **Error Handling**: Handle configuration fetch failures

### Operations

1. **Monitoring**: Monitor streaming query performance and resource usage
2. **Capacity Planning**: Use backend configuration to manage system load
3. **Gradual Rollout**: Deploy v4 backend incrementally
4. **Documentation**: Document backend-specific streaming behavior for users

## Future Enhancements

Potential future improvements to the streaming configuration system:

1. **Dynamic Configuration**: Real-time adjustment based on system conditions
2. **User-Specific Limits**: Different limits for different user roles
3. **Query Complexity Analysis**: Configuration based on query complexity
   metrics
4. **Caching Strategies**: Intelligent caching of streaming configuration
5. **Metrics and Monitoring**: Built-in metrics for streaming performance
