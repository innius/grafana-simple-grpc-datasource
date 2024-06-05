# Changelog

## 1.2.5

- feature: allow backend system to return string values

## 1.2.3

refactor: improved backend error messages feat: enhance width of metric
selection control in query editor chore: update to latest grafana framework fix:
reset query options after changing query type

## 1.2.2

bugfix: fix backend resource path

## 1.2.1

- feature: provide backend query options with default values
- feature: send currently selected query options to the backend while retrieving
  query options
- feature: include grafana time range in GetMetricQuery
- feature: layout improvements for query option editor

## 1.1.1.

- docs: explain differences between API versions

## 1.1.0

- feature: add new v3 API which allows backend to define query options

## 1.0.14

upgrade grafana dependencies and grafana tools

## 1.0.12

- feature: support backend value mapping

## 1.0.11

- fix: specify correct grafana version dependency

## 1.0.10

- feature: retry mechanism for backend grpc calls
- feature: add dimension value filter to VariableQueryEditor
- feature: update to latest grafana frameworks

## 1.0.9

- feature: imporoved dimension selection component
- feature: add variable editor which supports dimension keys and metrics

## 1.0.8

- bugfix: incorrect metrics when Explore is triggered from dashboard panel
- bugfix: fix dimension drop down glitch for Chrome browser

## 1.0.7

- feature: provide ListDimensions with selected dimensions to enable advanced
  backend filtering
- feature: support frame metadata to enable user notifications from backend
- bugfix: pagination for queries with multiple frames did not work properly

## 1.0.6

- upgrade: upgrade to latest grafana toolkit
- feature: add v2 API which better aligns with grafana dataframe API
- feature: add fieldname to display name expressions

## 1.0.5

- feature: add `COUNT` aggregate type to list of possible aggregations
- feature: support display name expression (also known as aliasing)
- bugfix: exclude variables from metricFind query

## 1.0.4

- improved metric selection
- support Count aggregation type
- add grafana `intervalMS` and `MaxItems` attributes to query definition

## 1.0.3

Hide technical grpc errors from user interface; backend plugin logs error
details and returns user-friendly message for the user.

## 1.0.2

- Add support for GetMetricAggregate query
- Fix a few typo's in Readme
- Correct plugin id to standard grafana plugin-id conventions

## 1.0.0 (Unreleased)

Initial release.
