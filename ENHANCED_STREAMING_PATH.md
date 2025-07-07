# Enhanced Streaming Path Composition

## Overview

The `createStreamingPath` method in `datasource.ts` has been enhanced to create more comprehensive unique identifiers for streaming queries. Previously, the path only included dimensions and the first metric, but now it includes all relevant query attributes.

## Changes Made

### 1. Enhanced Path Components

The streaming path now includes the following components in order:

1. **refId** - Query reference ID
2. **queryType** - Type of query (GetMetricHistory, GetMetricAggregate, GetMetricValue)
3. **metricId** - First metric ID (if available)
4. **dimensions** - All dimension key/value pairs
5. **queryOptions** - All query options with values
6. **streamingConfig** - Streaming configuration parameters

### 2. Path Format

```
/refId/queryType/metricId/dimensionKey1/dimensionValue1/dimensionKey2/dimensionValue2/optionKey1/optionValue1/buffer/bufferSize/lookback/lookbackPeriod
```

### 3. Example Paths

**Before Enhancement:**
```
A/temperature/location/room1/sensor/temp_01
```

**After Enhancement:**
```
A/GetMetricHistory/temperature/location/room1/sensor/temp_01/interval/5m/aggregation/avg/buffer/3600/lookback/300
```

### 4. Long Path Handling

For paths longer than 200 characters, the system automatically switches to a hash-based approach:

```
A/hashed/1a2b3c4d5e
```

This ensures URL length limits are respected while maintaining uniqueness.

### 5. Enhanced Query Display Text

The `getQueryDisplayText` method has also been enhanced to show:

- Query type in brackets: `[GetMetricHistory]`
- Dimensions: `[location=room1,sensor=temp_01]`
- Metrics: `temperature&pressure`
- Query options: `{interval=5m,aggregation=avg}`

Example: `[GetMetricHistory][location=room1,sensor=temp_01] temperature {interval=5m,aggregation=avg}`

## Benefits

### 1. Better Query Identification
- Each unique combination of query parameters gets a unique path
- Prevents conflicts between similar queries with different options

### 2. Improved Debugging
- Paths are more descriptive and easier to understand
- New `getStreamingPathDebugInfo()` method provides detailed path composition information

### 3. Streaming Configuration Support
- Different streaming configurations (buffer size, lookback period) get unique paths
- Enables multiple streaming queries with different configurations

### 4. Scalability
- Automatic hashing prevents URL length issues
- Maintains uniqueness even for complex queries

## Usage Examples

### Basic Query
```typescript
const query: MyQuery = {
  refId: 'A',
  queryType: QueryType.GetMetricHistory,
  metrics: [{ metricId: 'temperature' }],
  dimensions: [{ key: 'location', value: 'room1' }]
};
// Path: A/GetMetricHistory/temperature/location/room1
```

### Query with Options
```typescript
const query: MyQuery = {
  refId: 'B',
  queryType: QueryType.GetMetricAggregate,
  metrics: [{ metricId: 'pressure' }],
  dimensions: [{ key: 'location', value: 'room2' }],
  queryOptions: {
    aggregationType: { value: 'max', label: 'Maximum' }
  },
  streamingConfig: {
    maxBufferSize: 1800,
    lookBackPeriod: 600
  }
};
// Path: B/GetMetricAggregate/pressure/location/room2/aggregationType/max/buffer/1800/lookback/600
```

### Debug Information
```typescript
const debugInfo = datasource.getStreamingPathDebugInfo(query);
console.log(debugInfo);
// Output:
// {
//   path: "A/GetMetricHistory/temperature/location/room1/interval/5m",
//   components: {
//     refId: "A",
//     queryType: "GetMetricHistory",
//     firstMetric: "temperature",
//     dimensionsCount: 1,
//     queryOptionsCount: 1,
//     streamingConfig: { maxBufferSize: 3600, lookBackPeriod: 300 }
//   },
//   isHashed: false
// }
```

## Migration Notes

- Existing streaming queries will automatically use the new path format
- No breaking changes to the API
- Backward compatibility maintained through the enhanced path structure
- The new format provides better uniqueness and debugging capabilities

## Testing

Run the test script to see examples of the enhanced path creation:

```bash
node test-streaming-path.js
```

This will show how different query configurations generate unique paths and demonstrate the improvements in query identification.
