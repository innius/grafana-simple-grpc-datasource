# Grafana Simple gRPC Datasource Plugin

![GitHub release (latest by date)](https://img.shields.io/github/v/release/innius/grafana-simple-grpc-datasource?logo=github)
[![Marketplace](https://img.shields.io/badge/dynamic/json?logo=grafana&color=F47A20&label=marketplace&prefix=v&query=version&url=https%3A%2F%2Fgrafana.com%2Fapi%2Fplugins%2Finnius-grpc-datasource)](https://grafana.com/grafana/plugins/innius-grpc-datasource)
[![Downloads](https://img.shields.io/badge/dynamic/json?logo=grafana&color=F47A20&label=downloads&query=downloads&url=https%3A%2F%2Fgrafana.com%2Fapi%2Fplugins%2Finnius-grpc-datasource)](https://grafana.com/grafana/plugins/innius-grpc-datasource)

## What is this plugin?

This back-end Grafana datasource plugin provides a user-friendly grafana
experience with only a handful simple and generic parameters to configure. It
comes with a dedicated API specification that requires implementation in the
data provider's back-end. Implementing this API helps to decouple the front-end
visualisation solution from the back-end data-layer implementation, leaving
developers with the necessary freedom to update and improve the back-end without
breaking the end-user experience.

The protobuf API specification can be found in the pkg/proto directory. On
configuring the datasource plugin, the end-user provides an endpoint URL and
optionally an API key too. The datasource will attempt to establish a gRPC
connection and emit calls to the given endpoint according to the API
specification.

For more information on gRPC or protobuf, see the
[gRPC docs](https://grpc.io/docs/).

#### Why gRPC?

gRPC is a fast & efficient framework for inter-service communication and
provides a fool-proof and streamlined workflow for API implementation through
protobuf.

gRPC also supports all essential streaming capabilities, which can be
implemented in future releases.

#### Security

The datasource plugin establishes a secure gRPC connection through TLS.
Additionally, the datasource supports API-key authorization. The API-key will be
included in each API call as part of the call metadata.

## Usage

![screenshot](https://raw.githubusercontent.com/innius/grafana-simple-grpc-datasource/master/src/img/screenshots/image-1.png)

#### Metric

The variable that is updated with new values as the stream of timeseries
datapoints is appended.

#### Dimension

A dimension is an optional, identifying property of the measure. Each dimension
is modeled as a key-value pair. A measure can have zero or many dimensions that
collectively uniquely identify it.

#### Query Type

| type                 | description                       |
| -------------------- | --------------------------------- |
| Get Metric History   | gets historical timeseries values |
| Get Metric Aggregate | gets aggregated timeseries        |
| Get Metric Value     | gets the last known value         |

## Getting started

1. start a sample grpc server locally:

```
docker run -p 50051:50051 innius/sample-grpc-server
```

2. install the innius-simple-grpc-datasource

3. enable the datasource
   - configure the endpoint `localhost:50051`

4. configure dashboards

## Implement your own backend API

This datasource plugin expects a backend to implement the [Simple][1] or the
[Advanced][2] interface.

### The Simple API ([GrafanaQueryAPI][1])

This API provides the following operations:

| name                | description                                                         |
| ------------------- | ------------------------------------------------------------------- |
| ListDimensionKeys   | Returns a list of all available dimension keys                      |
| ListDimensionValues | Returns a list of all available dimension values of a dimension key |
| ListMetrics         | Returns a list of all metrics for a combination of dimensions.      |
| GetMetricValue      | Returns the last known value of a metric.                           |
| GetMetricHistory    | Returns historical values of a metric                               |
| GetMetricAggregate  | Returns aggregated metric values                                    |

A sample implementation can be found
[here](https://bitbucket.org/innius/sample-grpc-server/src/master/).

This API has some limitations:

- it only supports one metric per query
- it does not support variables with multiple options
- it does not support enhanced metadata for metrics (like unit, etc.)
- it does not support flexible query options

### The Advanced API ([GrafanaQueryAPIV3][3])

This API provides almost the same operations as the Simple API but with one
major difference: it supports multiple metrics for the same query. As a result
this API integrates seamlessly with grafana templating capabilities. In
addition, it supports enhanced metric metadata, like unit of measure. Another
difference is that it supports grafana labels.

The advanced API supports dynamic query options which are defined by the backend
system. This makes it possible to tailor the behavior of grafana queries for
specific backends. An example of a custom option is the Aggregate of the
_GetMetricAggregate_ query. The v1 version of the API has a fixed number of
Aggregates, defined by the plugin. It is not possible for a backend system to
add a different option. With the V3 API, however, this is supported. Currently
an option can be either an Enumeration or a Boolean type.

This API provides the following operations:

| name                | description                                                         |
| ------------------- | ------------------------------------------------------------------- |
| ListDimensionKeys   | Returns a list of all available dimension keys                      |
| ListDimensionValues | Returns a list of all available dimension values of a dimension key |
| ListMetrics         | Returns a list of all metrics for a combination of dimensions.      |
| GetMetricValue      | Returns the last known value for one or more metrics.               |
| GetMetricHistory    | Returns historical values for one or more metrics                   |
| GetMetricAggregate  | Returns aggregated values for one or more metrics                   |
| GetQueryOptions     | Returns the options for a selected query type                       |

A sample implementation can be found
[here](https://bitbucket.org/innius/sample-grpc-server/src/master/).

### The V4 API ([GrafanaQueryAPIV4][4])

The V4 API extends the V3 API with streaming configuration capabilities. It adds
one new operation:

| name                           | description                                                                         |
| ------------------------------ | ----------------------------------------------------------------------------------- |
| GetStreamingQueryConfiguration | Returns streaming configuration for a query (polling interval, max lookback period) |

This allows backend systems to control streaming behavior based on their
capabilities and constraints:

- **Polling Interval**: How frequently the datasource should poll for new data
  during streaming
- **Maximum Look-back Period**: The maximum allowed initial historical data
  period

The V4 API is fully backward compatible - existing V1/V2/V3 backends continue to
work with default streaming configuration.

For detailed information about V4 streaming configuration, see:

- [STREAMING_CONFIG.md](./STREAMING_CONFIG.md) - Technical documentation
- [V4_MIGRATION_GUIDE.md](./V4_MIGRATION_GUIDE.md) - Implementation guide for
  backend developers

## Differences Between V3 and V4 API

The V4 API is a minimal extension of the V3 API, focusing specifically on streaming configuration capabilities. Here are the key differences:

### New Features in V4

| Feature | V3 API | V4 API |
|---------|--------|--------|
| **Streaming Configuration** | ❌ Fixed defaults | ✅ Backend-controlled configuration |
| **Dynamic Polling Intervals** | ❌ Static intervals | ✅ Query-specific intervals via `LoopInterval` |
| **Configurable Look-back Period** | ❌ Fixed look-back | ✅ Query-specific look-back via `LookBackPeriod` |
| **Backward Compatibility** | N/A | ✅ Full V3 compatibility maintained |

### API Operations Comparison

#### V3 API Operations
- ListDimensionKeys
- ListDimensionValues  
- ListMetrics
- GetMetricValue
- GetMetricHistory
- GetMetricAggregate
- GetQueryOptions

#### V4 API Operations
- **All V3 operations** (unchanged)
- **GetStreamingQueryConfiguration** *(new)*

### Implementation Requirements

#### V3 Backend Requirements
- Implement all 7 V3 operations
- Support gRPC Reflection for API detection
- Handle multiple metrics per query
- Support custom query options

#### V4 Backend Requirements  
- **All V3 requirements** (unchanged)
- Implement `GetStreamingQueryConfiguration` method
- Return appropriate streaming configuration based on query type
- Handle streaming configuration requests for all query types

### Migration Path

#### For Backend Developers
1. **No Breaking Changes**: Existing V3 backends continue to work
2. **Optional Enhancement**: Add V4 streaming configuration support
3. **Simple Implementation**: Only one new method to implement

```go
func (s *YourServer) GetStreamingQueryConfiguration(
    ctx context.Context,
    req *v4.GetStreamingQueryConfigurationRequest,
) (*v4.GetStreamingQueryConfigurationResponse, error) {
    // Return streaming configuration based on query type
    return &v4.GetStreamingQueryConfigurationResponse{
        LookBackPeriod: durMillis(time.Hour),    // 1 hour lookback
        LoopInterval:   durMillis(time.Second),  // 1 second polling
    }, nil
}
```

#### For Frontend Users
- **No Changes Required**: Datasource automatically detects V4 capabilities
- **Enhanced Experience**: Better streaming performance when V4 backend is available
- **Fallback Support**: Graceful degradation to V3 behavior when V4 is not available

### Use Cases for V4

The V4 API is particularly beneficial for:

- **High-frequency data sources** that need faster polling intervals
- **Resource-constrained backends** that need longer polling intervals  
- **Different data types** requiring different look-back periods
- **Dynamic optimization** based on query complexity or data volume

### Detection and Fallback

The datasource uses gRPC reflection to automatically detect API version support:

1. **V4 Detection**: Checks for `GetStreamingQueryConfiguration` method
2. **Graceful Fallback**: Uses V3 behavior if V4 is not available
3. **No Configuration**: Users don't need to specify API version

#### Example Use Cases:

- different time series for the same metric with different labels. For example:
  the temperature measure is a room. The room has four zones: north, south, east
  and west. The V1 API does not support this unless there are four different
  metrics defined for each temperature / zone combination. The Advanced API does
  support this scenario by returning multiple time series for the same metric
  `temperature`, each annotated with different label `zone`.
- different time series for different metrics. For example: a room has multiple
  temperature sensors. The V1 API supports this by defining multiple queries for
  each metric. The Advanced API can do this with a single query.

Important Note: in order to use the Advanced API the backend server needs to
support [gRPC Reflection][3]. The plugin uses this to determine if a backend
supports the V2 or V3 protocol. If not supported it falls back on the Simple API
implementation.

Please note gRPC is programming language agnostic which makes it possible to
implement a backend in the language of your choice. Checkout the gRPC
[documentation](https://grpc.io/docs/languages/) of your language.

#### Changes between ([GrafanaQueryAPIV2][2]) and ([GravanaQueryAPIV3][3])

The most important difference is that the Aggregate types of the V2 API are not
available by the V3 API unless they are defined in the backend.

The backend code has to implement something like this:

```
const (
    // this id is important because it matches the current v2 aggregate type option 
	AggregationTypeOptionID = iota
    // these enum values are important because they match the values of the V2 options 
	AggregationTypeAverage = 0
	AggregationTypeMax     = 1
	AggregationTypeMin     = 2
	AggregationTypeCount   = 3
)


func (backend *BackendServerV3) GetQueryOptions(ctx context.Context, in *v3.GetOptionsRequest) (*v3.GetOptionsResponse, error) {
	var Options []*v3.Option
	switch in.GetQueryType() {
	case v3.GetOptionsRequest_GetMetricAggregate:
		Options = append(Options, []*v3.Option{
			{
				Id:          strconv.Itoa(AggregationTypeOptionID),
				Label:       "Aggregate",
				Description: "Aggregate the query results",
				Type:        v3.Option_Enum,
				EnumValues: []*v3.EnumValue{
					{Label: "Average", Description: "Calculate the average of the values", Id: strconv.Itoa(AggregationTypeAverage)},
					{Label: "Min", Description: "Calculate the minimum of the values", Id: strconv.Itoa(AggregationTypeMin)},
					{Label: "Max", Description: "Calculate the maximum of the values", Id: strconv.Itoa(AggregationTypeMax)},
					{Label: "Count", Description: "Calculate the sum of the values", Id: strconv.Itoa(AggregationTypeCount)},
				},
			},
		}...)
	case v3.GetOptionsRequest_GetMetricValue:
        return &v3.GetOptionsResponse{}, nil
	case v3.GetOptionsRequest_GetMetricHistory:
        return &v3.GetOptionsResponse{}, nil
	}
	return &v3.GetOptionsResponse{Options: Options}, nil
}
```

A sample implementation of the V3 backend can be found
[here](https://bitbucket.org/innius/sample-grpc-server/src/4dc9fd798eee92eb67c44085532e89518551a74d/server/v3/server.go#lines-44)

## Features

- select multiple metrics in one query
- flexible dimension selection
- integrated with Grafana variables and templating
- allow backend systems to provided additional metadata, like value mappings,
  unit of measure, etc.
- supports notifications
- supports pagination
- supports retries for grpc calls if backend server is at maximum capacity
- allow backend systems to define custom query options
- supports streaming queries with backend-controlled configuration (V4 API)

## Roadmap

- support annotations

[1]: https://raw.githubusercontent.com/innius/grafana-simple-grpc-datasource/master/pkg/proto/v1/api.proto
[2]: https://raw.githubusercontent.com/innius/grafana-simple-grpc-datasource/master/pkg/proto/v2/apiv2.proto
[3]: https://raw.githubusercontent.com/innius/grafana-simple-grpc-datasource/master/pkg/proto/v3/apiv3.proto
[4]: https://github.com/grpc/grpc/blob/master/doc/server-reflection.md
