# V4 Streaming Configuration Migration Guide

This guide helps you migrate your backend implementation to support the new v4 streaming configuration features.

## Quick Start

### For Backend Developers

1. **Add the v4 proto definition** to your project
2. **Implement GetStreamingQueryConfiguration** method
3. **Deploy your backend** - no client changes needed!

### For Frontend Users

No changes required! The datasource automatically detects and uses v4 capabilities when available.

## Backend Implementation

### Step 1: Add V4 Proto

Add the v4 proto file to your project and generate the gRPC code:

```bash
# Copy the v4 proto file
cp path/to/grafana-simple-grpc-datasource/pkg/proto/v4/apiv4.proto ./proto/

# Generate Go code (example)
protoc --go_out=. --go-grpc_out=. proto/apiv4.proto
```

### Step 2: Implement the Interface

Add the new method to your server implementation:

```go
package main

import (
    "context"
    "time"
    
    v4 "your-project/proto/v4"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type YourGrafanaServer struct {
    v4.UnimplementedGrafanaQueryAPIServer
    // ... your existing fields
}

// Implement the new streaming configuration method
func (s *YourGrafanaServer) GetStreamingQueryConfiguration(
    ctx context.Context,
    req *v4.GetStreamingQueryConfigurationRequest,
) (*v4.GetStreamingQueryConfigurationResponse, error) {
    
    // Analyze the query type and return appropriate configuration
    switch query := req.Query.(type) {
    case *v4.GetStreamingQueryConfigurationRequest_GetMetricValueRequest:
        return s.configureMetricValueStreaming(query.GetMetricValueRequest)
        
    case *v4.GetStreamingQueryConfigurationRequest_GetMetricHistoryRequest:
        return s.configureMetricHistoryStreaming(query.GetMetricHistoryRequest)
        
    case *v4.GetStreamingQueryConfigurationRequest_GetMetricAggregateRequest:
        return s.configureMetricAggregateStreaming(query.GetMetricAggregateRequest)
        
    default:
        return nil, status.Errorf(codes.InvalidArgument, "unsupported query type")
    }
}

// Example configuration methods
func (s *YourGrafanaServer) configureMetricValueStreaming(req *v4.GetMetricValueRequest) (*v4.GetStreamingQueryConfigurationResponse, error) {
    // For real-time metric values, allow frequent updates but limit history
    return &v4.GetStreamingQueryConfigurationResponse{
        LookBackPeriod: int64(30 * time.Minute / time.Millisecond), // 30 minutes max
        LoopInterval:   int64(5 * time.Second / time.Millisecond),   // Update every 5 seconds
    }, nil
}

func (s *YourGrafanaServer) configureMetricHistoryStreaming(req *v4.GetMetricHistoryRequest) (*v4.GetStreamingQueryConfigurationResponse, error) {
    // For historical data, allow longer lookback but slower updates
    return &v4.GetStreamingQueryConfigurationResponse{
        LookBackPeriod: int64(24 * time.Hour / time.Millisecond),    // 24 hours max
        LoopInterval:   int64(30 * time.Second / time.Millisecond),  // Update every 30 seconds
    }, nil
}

func (s *YourGrafanaServer) configureMetricAggregateStreaming(req *v4.GetMetricAggregateRequest) (*v4.GetStreamingQueryConfigurationResponse, error) {
    // For aggregates, moderate settings
    return &v4.GetStreamingQueryConfigurationResponse{
        LookBackPeriod: int64(6 * time.Hour / time.Millisecond),     // 6 hours max
        LoopInterval:   int64(15 * time.Second / time.Millisecond),  // Update every 15 seconds
    }, nil
}
```

### Step 3: Advanced Configuration

You can make configuration dynamic based on various factors:

```go
func (s *YourGrafanaServer) GetStreamingQueryConfiguration(
    ctx context.Context,
    req *v4.GetStreamingQueryConfigurationRequest,
) (*v4.GetStreamingQueryConfigurationResponse, error) {
    
    // Get base configuration
    baseConfig := s.getBaseConfigForQuery(req.Query)
    
    // Adjust based on system load
    if s.isHighLoad() {
        baseConfig.LoopInterval *= 2 // Slower updates during high load
        baseConfig.LookBackPeriod /= 2 // Less history during high load
    }
    
    // Adjust based on query complexity
    complexity := s.analyzeQueryComplexity(req.Query)
    if complexity > 0.8 { // High complexity
        baseConfig.LoopInterval *= 3 // Much slower updates for complex queries
    }
    
    // Adjust based on user/tenant
    userLimits := s.getUserLimits(ctx)
    if baseConfig.LookBackPeriod > userLimits.MaxLookBack {
        baseConfig.LookBackPeriod = userLimits.MaxLookBack
    }
    
    return baseConfig, nil
}

func (s *YourGrafanaServer) isHighLoad() bool {
    // Implement your load detection logic
    return s.currentCPUUsage > 0.8 || s.activeConnections > 1000
}

func (s *YourGrafanaServer) analyzeQueryComplexity(query interface{}) float64 {
    // Implement query complexity analysis
    // Consider factors like:
    // - Number of metrics
    // - Number of dimensions
    // - Time range
    // - Aggregation complexity
    return 0.5 // Example
}
```

## Configuration Strategies

### Strategy 1: Query Type Based

Different configurations for different query types:

```go
var streamingConfigs = map[string]*v4.GetStreamingQueryConfigurationResponse{
    "MetricValue": {
        LookBackPeriod: int64(15 * time.Minute / time.Millisecond),
        LoopInterval:   int64(2 * time.Second / time.Millisecond),
    },
    "MetricHistory": {
        LookBackPeriod: int64(12 * time.Hour / time.Millisecond),
        LoopInterval:   int64(30 * time.Second / time.Millisecond),
    },
    "MetricAggregate": {
        LookBackPeriod: int64(6 * time.Hour / time.Millisecond),
        LoopInterval:   int64(15 * time.Second / time.Millisecond),
    },
}
```

### Strategy 2: Metric-Specific Configuration

Different configurations for different types of metrics:

```go
func (s *YourGrafanaServer) configureByMetricType(metrics []string) *v4.GetStreamingQueryConfigurationResponse {
    // Check if any metrics are high-frequency
    hasHighFrequencyMetrics := false
    for _, metric := range metrics {
        if s.isHighFrequencyMetric(metric) {
            hasHighFrequencyMetrics = true
            break
        }
    }
    
    if hasHighFrequencyMetrics {
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: int64(5 * time.Minute / time.Millisecond),  // Short history
            LoopInterval:   int64(1 * time.Second / time.Millisecond),   // Fast updates
        }
    }
    
    // Default configuration for regular metrics
    return &v4.GetStreamingQueryConfigurationResponse{
        LookBackPeriod: int64(1 * time.Hour / time.Millisecond),
        LoopInterval:   int64(10 * time.Second / time.Millisecond),
    }
}
```

### Strategy 3: Resource-Aware Configuration

Adjust configuration based on available resources:

```go
func (s *YourGrafanaServer) getResourceAwareConfig() *v4.GetStreamingQueryConfigurationResponse {
    memUsage := s.getMemoryUsage()
    cpuUsage := s.getCPUUsage()
    
    // Base configuration
    config := &v4.GetStreamingQueryConfigurationResponse{
        LookBackPeriod: int64(2 * time.Hour / time.Millisecond),
        LoopInterval:   int64(10 * time.Second / time.Millisecond),
    }
    
    // Adjust based on resource usage
    if memUsage > 0.8 {
        config.LookBackPeriod /= 2 // Reduce memory usage
    }
    
    if cpuUsage > 0.8 {
        config.LoopInterval *= 2 // Reduce CPU usage
    }
    
    return config
}
```

## Testing Your Implementation

### Unit Tests

```go
func TestGetStreamingQueryConfiguration(t *testing.T) {
    server := &YourGrafanaServer{}
    
    tests := []struct {
        name     string
        request  *v4.GetStreamingQueryConfigurationRequest
        expected *v4.GetStreamingQueryConfigurationResponse
    }{
        {
            name: "MetricValue query",
            request: &v4.GetStreamingQueryConfigurationRequest{
                Query: &v4.GetStreamingQueryConfigurationRequest_GetMetricValueRequest{
                    GetMetricValueRequest: &v4.GetMetricValueRequest{
                        Metrics: []string{"cpu.usage"},
                    },
                },
            },
            expected: &v4.GetStreamingQueryConfigurationResponse{
                LookBackPeriod: int64(30 * time.Minute / time.Millisecond),
                LoopInterval:   int64(5 * time.Second / time.Millisecond),
            },
        },
        // Add more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := server.GetStreamingQueryConfiguration(context.Background(), tt.request)
            assert.NoError(t, err)
            assert.Equal(t, tt.expected.LookBackPeriod, resp.LookBackPeriod)
            assert.Equal(t, tt.expected.LoopInterval, resp.LoopInterval)
        })
    }
}
```

### Integration Tests

```go
func TestStreamingConfigurationIntegration(t *testing.T) {
    // Start your gRPC server
    server := startTestServer()
    defer server.Stop()
    
    // Connect with v4 client
    conn, err := grpc.Dial(server.Address(), grpc.WithInsecure())
    require.NoError(t, err)
    defer conn.Close()
    
    client := v4.NewGrafanaQueryAPIClient(conn)
    
    // Test the streaming configuration
    req := &v4.GetStreamingQueryConfigurationRequest{
        Query: &v4.GetStreamingQueryConfigurationRequest_GetMetricHistoryRequest{
            GetMetricHistoryRequest: &v4.GetMetricHistoryRequest{
                Metrics: []string{"memory.usage"},
            },
        },
    }
    
    resp, err := client.GetStreamingQueryConfiguration(context.Background(), req)
    require.NoError(t, err)
    assert.Greater(t, resp.LookBackPeriod, int64(0))
    assert.Greater(t, resp.LoopInterval, int64(0))
}
```

## Deployment

### Gradual Rollout

1. **Deploy v4 Backend**: Your backend now supports both v3 and v4 APIs
2. **Monitor**: Check logs to see v4 API usage
3. **Optimize**: Adjust configuration based on real usage patterns

### Monitoring

Add monitoring to track streaming configuration usage:

```go
func (s *YourGrafanaServer) GetStreamingQueryConfiguration(
    ctx context.Context,
    req *v4.GetStreamingQueryConfigurationRequest,
) (*v4.GetStreamingQueryConfigurationResponse, error) {
    
    // Metrics collection
    s.streamingConfigRequests.Inc()
    
    config, err := s.calculateStreamingConfig(req)
    if err != nil {
        s.streamingConfigErrors.Inc()
        return nil, err
    }
    
    // Log configuration for analysis
    s.logger.Info("Streaming configuration provided",
        "queryType", getQueryType(req.Query),
        "lookBackPeriod", config.LookBackPeriod,
        "loopInterval", config.LoopInterval,
    )
    
    return config, nil
}
```

## Common Patterns

### Pattern 1: Tiered Service Levels

```go
type ServiceLevel int

const (
    Basic ServiceLevel = iota
    Premium
    Enterprise
)

func (s *YourGrafanaServer) getConfigForServiceLevel(level ServiceLevel) *v4.GetStreamingQueryConfigurationResponse {
    switch level {
    case Basic:
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: int64(1 * time.Hour / time.Millisecond),
            LoopInterval:   int64(30 * time.Second / time.Millisecond),
        }
    case Premium:
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: int64(6 * time.Hour / time.Millisecond),
            LoopInterval:   int64(10 * time.Second / time.Millisecond),
        }
    case Enterprise:
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: int64(24 * time.Hour / time.Millisecond),
            LoopInterval:   int64(5 * time.Second / time.Millisecond),
        }
    default:
        return s.getDefaultConfig()
    }
}
```

### Pattern 2: Time-Based Configuration

```go
func (s *YourGrafanaServer) getTimeBasedConfig() *v4.GetStreamingQueryConfigurationResponse {
    hour := time.Now().Hour()
    
    // During business hours (9 AM - 5 PM), allow more frequent updates
    if hour >= 9 && hour <= 17 {
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: int64(2 * time.Hour / time.Millisecond),
            LoopInterval:   int64(5 * time.Second / time.Millisecond),
        }
    }
    
    // During off-hours, reduce frequency
    return &v4.GetStreamingQueryConfigurationResponse{
        LookBackPeriod: int64(1 * time.Hour / time.Millisecond),
        LoopInterval:   int64(30 * time.Second / time.Millisecond),
    }
}
```

### Pattern 3: Adaptive Configuration

```go
type AdaptiveConfig struct {
    baseConfig     *v4.GetStreamingQueryConfigurationResponse
    loadThreshold  float64
    adaptationRate float64
}

func (ac *AdaptiveConfig) GetConfig(currentLoad float64) *v4.GetStreamingQueryConfigurationResponse {
    if currentLoad <= ac.loadThreshold {
        return ac.baseConfig
    }
    
    // Adapt configuration based on load
    adaptationFactor := 1.0 + (currentLoad-ac.loadThreshold)*ac.adaptationRate
    
    return &v4.GetStreamingQueryConfigurationResponse{
        LookBackPeriod: int64(float64(ac.baseConfig.LookBackPeriod) / adaptationFactor),
        LoopInterval:   int64(float64(ac.baseConfig.LoopInterval) * adaptationFactor),
    }
}
```

## Troubleshooting

### Issue: Configuration Not Applied

**Symptoms**: Client still uses default configuration

**Solutions**:
1. Check if v4 API is properly implemented
2. Verify gRPC reflection is enabled
3. Check server logs for errors
4. Test the endpoint directly with grpcurl

```bash
# Test the endpoint
grpcurl -plaintext -d '{
  "query": {
    "getMetricValueRequest": {
      "metrics": ["cpu.usage"]
    }
  }
}' localhost:50051 grafanav4.GrafanaQueryAPI/GetStreamingQueryConfiguration
```

### Issue: High Resource Usage

**Symptoms**: Backend overloaded with streaming requests

**Solutions**:
1. Increase polling intervals during high load
2. Reduce maximum lookback periods
3. Implement rate limiting
4. Add circuit breakers

### Issue: Client Preferences Ignored

**Symptoms**: Client lookback period not respected

**Solutions**:
1. Check if client preference exceeds backend maximum
2. Verify duration format (use "1h", "30m", "5s")
3. Check frontend validation logic
4. Review backend maximum values

## Best Practices Summary

1. **Start Conservative**: Begin with longer intervals and shorter lookback periods
2. **Monitor Performance**: Track resource usage and adjust accordingly
3. **User Experience**: Balance real-time updates with system performance
4. **Graceful Degradation**: Handle high load scenarios gracefully
5. **Documentation**: Document your configuration strategy for users
6. **Testing**: Thoroughly test different load scenarios
7. **Monitoring**: Implement comprehensive monitoring and alerting

## Next Steps

After implementing v4 streaming configuration:

1. **Monitor Usage**: Track how the configuration affects system performance
2. **Gather Feedback**: Collect user feedback on streaming behavior
3. **Optimize**: Fine-tune configuration based on real-world usage
4. **Scale**: Plan for increased streaming usage
5. **Enhance**: Consider additional configuration parameters for future versions

## Support

If you encounter issues during migration:

1. Check the [V4_STREAMING_CONFIG.md](./V4_STREAMING_CONFIG.md) for detailed technical information
2. Review the example implementation in the sample server
3. Enable debug logging to troubleshoot configuration issues
4. Test with a simple configuration first, then add complexity

Remember: The v4 streaming configuration is backward compatible. Existing v1/v2/v3 backends continue to work with default configuration, so you can migrate at your own pace.
