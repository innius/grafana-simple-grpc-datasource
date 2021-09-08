# 2. add support multi metric queries

Date: 2021-09-07

## Status

Accepted

## Context

One of the most popular grafana features are [templates and variables](https://grafana.com/docs/grafana/latest/variables/). Variables allow you to create more interactive and dynamic reports. The initial version of this plugin supports variables but is limited to only one selected variable. 
Removing this limitation improves the usability of this plugin significantly: 
* it becomes possible to select multiple metrics per panel query.
* performance improvement because multiple queries can be combined in one network call. 

Making such a change means a breaking change for both the request and the responses of all queries. This document describes various options for implementing such a change.

## Decision Drivers 

* Migration; a lock-step deployment in which all dependencies have to be updated at the same time is unacceptable. It should be possible to roll out the change in a gradual manner. 
* Ease of use; the api should be intuitive to use.

## Options 

### Option 1: break the API 

* Change the metric specification from a single string value to an array.
* Change the definition of a response value to include the metric identification 
* Change the response from an array of values into an array of metric values

#### For example: GetMetricHistory 

```protobuf
  // Gets the history of a metric's values
  rpc GetMetricHistory (GetMetricHistoryRequest) returns (GetMetricHistoryResponse) { }
```
Original Request: 
```protobuf
message GetMetricHistoryRequest {
    repeated Dimension dimensions = 3;
    // a single metric 
    string Metric = 4;
    int64 startDate = 5;
    int64 endDate = 6;
    int64 maxItems = 7;
    TimeOrdering timeOrdering = 8;
    string startingToken = 9;
}
```

Modified Request:
```protobuf
message GetMetricHistoryRequest {
  repeated Dimension dimensions = 3;
  // an array of metrics 
  repeated Metric Metrics = 4;
  int64 startDate = 5;
  int64 endDate = 6;
  int64 maxItems = 7;
  TimeOrdering timeOrdering = 8;
  string startingToken = 9;
}
```

Original Response: 
```protobuf
message GetMetricHistoryResponse {
    // The asset property's value history.
    repeated MetricHistoryValue values = 1;

    // The token for the next set of results, or null if there are no additional results.
    string nextToken = 2;
}
```

Modified Response: 
```protobuf
// the timestamped values of a selected metric
message MetricValues {
  Metric metric = 1;
  repeated MetricValue values = 2;
}

// a single metric timestamped value
message MetricValue{
  // The timestamp date, in seconds, in the Unix epoch format.
  int64 timestamp = 1;

  // The current metric value.
  Value value = 2;
}

message GetMetricAggregateResponse {
  // The metric value history.
  repeated MetricValues values = 1;

  // The token for the next set of results, or null if there are no additional results.
  string nextToken = 2;
}
```

**Pros:**
* Clear API

**Cons:**
* Breaking change -> impacts plugin and backend api's
* Do not allow gradual migration model


### Option 2: keep interface intact and support comma separated strings 

This option does not break the API but specifies metrics as comma separated strings. It is up to the backend implementation to parse this string or not. 

**Pros**:

* Existing interfaces do not break 
* Only the backend API needs to change
* Allows a more gradual migration model 

**Cons**:

* Not an intuitive API; Metric is a simple string but contains either single value or an array.
* Response types do not explicitly support this (metric id not included in the response)
* Will be more difficult to maintain in the future 

### Option 3: enhance the interface with new operations 

This option adds new batch operations to the API. 

For example: GetMetricHistory

```protobuf
// deprecated 
rpc GetMetricHistory (GetMetricHistoryRequest) returns (GetMetricHistoryResponse) { } 

// Gets the history of a metric's values
rpc GetMetricHistoryV2 (GetMetricHistoryV2Request) returns (GetMetricHistoryV2Response) { }

message GetMetricHistoryV2Request {
  repeated Dimension dimensions = 3;
  // an array of metrics 
  repeated Metric Metrics = 4;
  int64 startDate = 5;
  int64 endDate = 6;
  int64 maxItems = 7;
  TimeOrdering timeOrdering = 8;
  string startingToken = 9;
}

message GetMetricHistoryV2Response {
    // The asset property's value history.
    repeated MetricHistoryValue values = 1;

    // The token for the next set of results, or null if there are no additional results.
    string nextToken = 2;
}

// the timestamped values of a selected metric
message MetricValues {
  Metric metric = 1;
  repeated MetricValue values = 2;
}

// a single metric timestamped value
message MetricValue{
  // The timestamp date, in seconds, in the Unix epoch format.
  int64 timestamp = 1;

  // The current metric value.
  Value value = 2;
}
```

**Pros:**
* Existing interfaces do not break 
* Supports gradual migration model: allows deprecating existing operations

**Cons:**
* Complexity of the API increases (until removal of deprecated operations)

### Option 4: enhance the current API 

This option enhances the current API with a new field for selecting multiple metrics. 

**Pros:** 
* No breaking change 
* Supports gradual migration model

**Cons:**
* Complexity of the API increases, especially for the response types. 
* Current response types do not support this
* Will become more difficult to implement because of different implementations of the same operation

#### Option 5: add a new version of the API 

Create a new version of the API without altering the current: create a copy of the current `.proto` file and change the API as required. As a result a new go package is generated for the v2 version of the API.

Current API can be deprecated in favor of the new version. Backend implementations can migrate to the latest version.

This option supports the changes proposed in option 1. 

Grpc backend implementations support multiple versions of the same API. 

**Pros:**
* No breaking change 
* Supports gradual migration model
* Allow refactoring without cluttering the API with deprecated methods
* Great level of flexibility

**Cons:**
* Increased complexity Backend API needs to support two versions of the API 


## Decision
Option 5 seems to be the best option because it does not break anything and allows us to roll out this option gradually. 

## Consequences

* Define a new v2 version of the API which supports multiple metrics
* Inform consumers to migrate their API's to the v2 version of the API 
* Deprecate current version of the API 
* Remove current version of the API if all consumers are migrated. 
