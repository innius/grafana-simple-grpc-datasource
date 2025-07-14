# V4 Streaming Configuration Migration Guide

This guide helps you migrate your backend implementation to support the new v4
streaming configuration features.

## Quick Start

### For Backend Developers

1. **Add the v4 proto definition** to your project
2. **Implement GetStreamingQueryConfiguration** method
3. **Deploy your backend** - no client changes needed!

### For Frontend Users

No changes required! The datasource automatically detects and uses v4
capabilities when available.

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
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: durMillis(time.Hour),
            LoopInterval:   durMillis(time.Second),
        }, nil
    case *v4.GetStreamingQueryConfigurationRequest_GetMetricHistoryRequest:
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: durMillis(time.Hour),
            LoopInterval:   durMillis(time.Second),
        }, nil
    case *v4.GetStreamingQueryConfigurationRequest_GetMetricAggregateRequest:
        return &v4.GetStreamingQueryConfigurationResponse{
            LookBackPeriod: durMillis(time.Hour),
            LoopInterval:   durMillis(time.Second),
        }, nil
    default:
        return nil, status.Errorf(codes.InvalidArgument, "unsupported query type")
    }
}

func durMillis(d time.Duration) int64 {
    return d.Milliseconds()
}
```
