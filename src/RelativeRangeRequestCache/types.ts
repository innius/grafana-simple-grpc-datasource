import { DataFrame } from '@grafana/data';
import { QueryType, MyQuery } from 'types';

const TIME_SERIES_QUERY_TYPES = new Set<QueryType>([
  QueryType.GetMetricValue,
  QueryType.GetMetricHistory,
  QueryType.GetMetricAggregate,
]);

export function isTimeSeriesQueryType(queryType: QueryType) {
  return TIME_SERIES_QUERY_TYPES.has(queryType);
}

const TIME_ORDERING_QUERY_TYPES = new Set<QueryType>([QueryType.GetMetricHistory, QueryType.GetMetricAggregate]);

export function isTimeOrderingQueryType(queryType: QueryType) {
  return TIME_ORDERING_QUERY_TYPES.has(queryType);
}

export interface CachedQueryInfo {
  query: MyQuery;
  dataFrame: DataFrame;
}
