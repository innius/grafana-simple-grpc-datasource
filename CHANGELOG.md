# Changelog

## 1.0.7 
* feature: provide ListDimensions with selected dimensions to enable advanced backend filtering
* feature: support frame metadata to enable user notifications from backend 
* bugfix: pagination for queries with multiple frames did not work properly 

## 1.0.6
* upgrade: upgrade to latest grafana toolkit
* feature: add v2 API which better aligns with grafana dataframe API 
* feature: add fieldname to display name expressions

## 1.0.5
* feature: add `COUNT` aggregate type to list of possible aggregations
* feature: support display name expression (also known as aliasing)
* bugfix: exclude variables from metricFind query

## 1.0.4

* improved metric selection 
* support Count aggregation type
* add grafana `intervalMS` and `MaxItems` attributes to query definition 

## 1.0.3

Hide technical grpc errors from user interface; backend plugin logs error details and returns user-friendly message for the user.

## 1.0.2

- Add support for GetMetricAggregate query
- Fix a few typo's in Readme
- Correct plugin id to standard grafana plugin-id conventions

## 1.0.0 (Unreleased)

Initial release.
