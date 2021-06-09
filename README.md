# Grafana Simple gRPC Datasource Plugin

## What is this plugin?

This back-end Grafana datasource plugin provides a user-friendly grafana experience with only a handful simple and generic parameters to configure.
It comes with a dedicated API specification that requires implementation in the data provider's back-end.
Implementing this API helps to decouple the front-end visualisation solution from the back-end data-layer implementation,
leaving developers with the necessary freedom to update and improve the back-end without breaking the end-user experience.

The protobuf API specification can be found in the pkg/proto directory.
On configuring the datasource plugin, the end-user provides an endpoint URL and optionally an API key too. The datasource will
attempt to establish a gRPC connection and emit calls to the given endpoint according to the API specification.

For more information on gRPC or protobuf, see the [gRPC docs](https://grpc.io/docs/).

#### Why gRPC?
gRPC is a fast & efficient framework for inter-service communication and provides a fool-proof and streamlined workflow for API implementation through protobuf.

gRPC also supports all essential streaming capabilities, which can be implemented in future releases.

#### Security

The datasource plugin establishes a secure gRPC connection through TLS. 
Additionally, the datasource supports API-key authorization. The API-key will be included in each API call as part of the call metadata.

##  Usage
![screenshot](https://raw.githubusercontent.com/innius/grafana-simple-grpc-datasource/master/src/img/screenshots/image-1.png)
#### Metric
The variable that is updated with new values as the stream of timeseries datapoints is appended.

#### Dimension
A dimension is an optional, identifying property of the measure. Each dimension is modeled as a key-value pair. 
A measure can have zero or many dimensions that collectively uniquely identify it.

#### Query Type

| type | description |
| --- | --- |
| Get Metric History | gets historical timeseries values of a metric for the selected period |
| Get Metric Value | gets the current value or last known value of a specified metric.  


## Getting started
1. start a sample grpc server locally:
```
docker run -p 50051:50041 innius/sample-grpc-server
```
   
2. install the innius-simple-grpc-datasource

3. enable the datasource 
    - configure the endpoint `localhost:50051`
    
4. configure dashboards 

## Implement your own backend API 

This datasource plugin expects a backend to implement `GrafanaQueryAPI` interface. The definition of this interface can be found [here](https://raw.githubusercontent.com/innius/grafana-simple-grpc-datasource/master/pkg/proto/api.proto). This API provides the following operations:

| name | description | 
| --- | --- |
| ListDimensionKeys| Returns a list of all available dimension keys |
| ListDimensionValues | Returns a list of all available dimension values of a dimension key |
| ListMetrics | Returns a list of all metrics for a combination of dimensions. |
| GetMetricValue | Returns the last known value of a metric. |
| GetMetricHistory | Returns historical values of a metric |

A sample implementation can be found [here](https://bitbucket.org/innius/sample-grpc-server/src/master/).

Please note gRPC is programming language agnostic which makes it possible to implement a backend in the language of your choice. Checkout the gRPC [documentation](https://grpc.io/docs/languages/) of your language.

## Roadmap
- [ ] add pagination to historical value query 
- [ ] add more caching 
- [ ] add more authentication schemes (certificates, basic authentication etc. )
- [ ] add more tests 
- [ ] better lookups for dimensions and metrics in frontend 
- [ ] support annotations 
- [ ] support for aggregations
- [ ] support streaming queries 
