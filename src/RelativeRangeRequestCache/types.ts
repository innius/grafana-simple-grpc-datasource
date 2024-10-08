import { DataFrame } from '@grafana/data';
import { QueryType, MyQuery, GetMetricHistoryQuery } from 'types';

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
  query: SitewiseQueriesUnion;
  dataFrame: DataFrame;
}

// Union of all SiteWise queries variants
export type SitewiseQueriesUnion = MyQuery &
  // Partial<Pick<GetMetricAggregateQuery, 'aggregates'>> &
  Partial<Pick<GetMetricHistoryQuery, 'timeOrdering'>>;
// Partial<Pick<ListAssociatedAssetsQuery, 'loadAllChildren'>> &
// Partial<Pick<ListAssociatedAssetsQuery, 'hierarchyId'>> &
// Partial<Pick<ListAssetsQuery, 'modelId'>> &
// Partial<Pick<ListAssetsQuery, 'filter'>> &
// Partial<Pick<ListTimeSeriesQuery, 'timeSeriesType'>> &
// Partial<Pick<ListTimeSeriesQuery, 'aliasPrefix'>>;
