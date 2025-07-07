# Streaming Configuration

This document describes the enhanced streaming configuration feature for the Grafana Simple gRPC Datasource Plugin.

## Overview

The streaming configuration feature provides fine-grained control over how streaming queries behave, allowing users to configure:

- **Streaming Toggle**: Enable/disable streaming using boolean values or Grafana template variables
- **Max Buffer Size**: Maximum number of datapoints to keep in the streaming buffer
- **Look-back Period**: Initial historical data period to load when starting the stream

## Streaming Control Options

### Boolean Mode (Default)
- Simple on/off toggle switch
- Direct boolean value (true/false)
- Immediate visual feedback

### Variable Mode
- Text input field for Grafana template variables
- Supports variable references like `$streaming`, `${streaming}`
- Supports static string values like `"true"`, `"false"`
- Dynamic control via dashboard variables

### Mode Switching
Users can switch between modes using the toggle button:
- **"Use Variable"** button: Switch from boolean to variable mode
- **"Switch to Toggle"** button: Switch from variable to boolean mode

## Configuration Options

### Streaming Control

| Mode | Input Type | Example Values | Use Case |
|------|------------|----------------|----------|
| **Boolean** | Toggle Switch | `true`, `false` | Simple on/off control |
| **Variable** | Text Input | `$streaming`, `${enable_streaming}`, `"true"` | Dashboard-wide control |

### Max Buffer Size

- **Purpose**: Controls the maximum number of datapoints kept in memory for the streaming buffer
- **Default**: 3600 datapoints
- **Range**: 100 - 100,000 datapoints
- **Impact**: 
  - Higher values provide more historical context but use more memory
  - Lower values use less memory but provide less historical data for visualization

### Look-back Period

- **Purpose**: Defines how much historical data (in seconds) to load when the stream starts
- **Default**: 300 seconds (5 minutes)
- **Range**: 0 - 86400 seconds (24 hours)
- **Impact**:
  - 0 seconds: Start streaming with no historical data
  - Higher values: Load more historical context before streaming begins

## User Interface

### Streaming Toggle
The streaming control appears as either:
1. **Toggle Switch** (Boolean Mode): Simple on/off switch
2. **Text Input** (Variable Mode): Text field for variables or static values

### Streaming Configuration Panel
When streaming is enabled, a "Streaming Configuration" panel appears with:

1. **Max Buffer Size** input field
   - Number input with step increment of 100
   - Tooltip explaining memory usage implications

2. **Look-back Period** input field  
   - Number input with step increment of 60 seconds
   - Tooltip explaining initial data loading behavior

3. **Help text** explaining the impact of both settings

## Variable Usage Examples

### Dashboard-Wide Streaming Control

1. **Create a Dashboard Variable**:
   - Name: `streaming`
   - Type: `Custom`
   - Values: `true,false`
   - Current value: `true`

2. **Use in Query**:
   - Switch to Variable Mode
   - Enter: `$streaming`
   - All queries will respect the dashboard variable

### Environment-Based Control

1. **Create Environment Variable**:
   - Name: `environment`
   - Type: `Custom`
   - Values: `development,staging,production`

2. **Use Conditional Logic**:
   - Enter: `${environment:regex:/(development|staging)/}`
   - Streaming enabled only for dev/staging environments

### Static Configuration

- Enter `"true"` or `"false"` as static strings
- Useful for template dashboards or configuration management

## Technical Implementation

### Type Definitions

```typescript
export interface StreamingConfig {
  // Maximum number of datapoints to keep in buffer
  maxBufferSize?: number;
  // Look-back period in seconds for initial dataset
  lookBackPeriod?: number;
}

export interface MyQuery extends DataQuery {
  // ... other properties
  // Streaming can be boolean or string (for variables)
  isStreaming?: boolean | string;
  streamingConfig?: StreamingConfig;
}
```

### Variable Resolution

The datasource resolves streaming variables during query execution:

```typescript
// Template variable resolution
const resolvedIsStreaming = templateSrv.replace(query.isStreaming, scopedVars);

// Boolean evaluation
private isStreamingEnabled(query: MyQuery): boolean {
  if (typeof query.isStreaming === 'boolean') return query.isStreaming;
  if (typeof query.isStreaming === 'string') {
    const lowerValue = query.isStreaming.toLowerCase().trim();
    return lowerValue === 'true' || lowerValue === '1' || lowerValue === 'yes';
  }
  return false;
}
```

### Default Values

```typescript
export const defaultQuery: Partial<MyQuery> = {
  // ... other defaults
  streamingConfig: {
    maxBufferSize: 3600,
    lookBackPeriod: 300,
  },
};
```

## Usage Examples

### High-Frequency Data with Large Buffer

For high-frequency sensor data where you need to see trends over time:

```
Streaming: $streaming (dashboard variable)
Max Buffer Size: 10000
Look-back Period: 1800 (30 minutes)
```

### Environment-Specific Streaming

Enable streaming only in development:

```
Streaming: ${environment:regex:/development/}
Max Buffer Size: 3600
Look-back Period: 300
```

### Memory-Constrained Environment

For environments with limited memory:

```
Streaming: true
Max Buffer Size: 500
Look-back Period: 60 (1 minute)
```

### Real-time Monitoring Dashboard

Dashboard with streaming toggle variable:

1. Create variable `streaming_enabled` with values `true,false`
2. Set streaming field to `$streaming_enabled`
3. Users can toggle streaming for entire dashboard

## Best Practices

### Variable Management
1. **Consistent Naming**: Use consistent variable names across dashboards
2. **Default Values**: Set sensible defaults for variables
3. **Documentation**: Document variable purposes in dashboard descriptions

### Performance Optimization
1. **Memory Considerations**: Monitor memory usage when increasing buffer size
2. **Conditional Streaming**: Use variables to enable streaming only when needed
3. **Environment Awareness**: Different settings for dev/staging/production

### User Experience
1. **Clear Labels**: Use descriptive variable labels and help text
2. **Reasonable Defaults**: Provide good default values for different scenarios
3. **Progressive Disclosure**: Show advanced options only when needed

## Migration

### From Boolean to Variable Mode
Existing queries with boolean `isStreaming` values continue to work unchanged. Users can switch to variable mode as needed.

### Backwards Compatibility
- Boolean values: Fully supported
- String values: Evaluated as described above
- Missing values: Default to `false`

No migration is required as the feature is fully backwards compatible.

## Testing

The streaming configuration includes comprehensive unit tests covering:
- Boolean and variable mode rendering
- Mode switching functionality
- Variable evaluation logic
- Configuration change handling
- Input validation
- Disabled state behavior

Run tests with:
```bash
npm test -- --testPathPattern=StreamingToggle
npm test -- --testPathPattern=StreamingConfigEditor
```
