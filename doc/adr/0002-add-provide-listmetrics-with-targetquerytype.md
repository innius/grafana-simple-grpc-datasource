# 2. add provide ListMetrics with target QueryType

Date: 2021-12-01

## Status

Rejected

## Context

Some metrics do not support all query types. The list of metrics, however, does not have any context about the query for which query type the metrics are selected. 

Providing the `ListMetrics` API with more context information would make it possible to filter metric lists

## Solution: add target query type to `ListMetrics` API

Extend the ListMetrics  API with an additional attributes which specifies the query type of the current query. The backend can use this property to filter selected metrics. 

*Pros:*
* easy to implemented
* allows backend to filter metrics which support the selected query type only

*Cons:*
* inconsistent user experience; from a user perspective it is not clear why certain metrics are not included in the list. This is a degradation in user experience.
* does not work properly with grafana variables; The scope of a variable is the dashboard and therefore there is no information about the current query.   

## Decision

Reject this change because it degrades the user experience of the plugin. 

## Consequences

ListMetrics always returns all metrics, regardless the selected query type. 
