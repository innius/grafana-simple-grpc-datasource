import { DataQuery, DataSourceJsonData } from '@grafana/data';

export enum QueryType {
  ListDimensionKeys = 'ListDimensionKeys',
  ListDimensionValues = 'ListDimensionValues',
  ListMetrics = 'ListMetrics',
  GetMetricValue = 'GetMetricValue',
  GetMetricHistory = 'GetMetricHistory',
  GetMetricAggregate = 'GetMetricAggregate',
}

export enum AggregateType {
  AVERAGE = 'AVERAGE',
  // COUNT = 'COUNT',
  MAXIMUM = 'MAXIMUM',
  MINIMUM = 'MINIMUM',
  // SUM = 'SUM',
  // STANDARD_DEVIATION = 'STANDARD_DEVIATION',
}

export function isMetricQuery(queryType: QueryType): boolean {
  return (
    queryType === QueryType.GetMetricValue ||
    queryType === QueryType.GetMetricHistory ||
    queryType === QueryType.GetMetricAggregate
  );
}

export interface MyQuery extends DataQuery {
  queryType: QueryType;
  dimensions?: Dimensions;
  metricName?: string;
  metricId?: string;
  aggregateType?: AggregateType;
}

export interface Dimension {
  id: string;
  key: string;
  value: string;
}

export type Dimensions = Dimension[];

export const defaultQuery: Partial<MyQuery> = {
  dimensions: [],
};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  endpoint?: string;
  apikey_authentication_enabled: boolean;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  apiKey?: string;
}

export interface GetMetricValueQuery extends MyQuery {
  queryType: QueryType.GetMetricValue;
}

export interface GetMetricHistoryQuery extends MyQuery {
  queryType: QueryType.GetMetricHistory;
}

export interface GetMetricAggregateQuery extends MyQuery {
  queryType: QueryType.GetMetricAggregate;
}

export interface ListDimensionsQuery extends MyQuery {
  queryType: QueryType.ListDimensionKeys;
  filter: string;
}

export interface ListDimensionValuesQuery extends MyQuery {
  queryType: QueryType.ListDimensionValues;
  dimensionKey: string;
  filter: string;
}

export interface ListMetricsQuery extends MyQuery {
  queryType: QueryType.ListMetrics;
  dimensions: Dimensions;
  filter: string;
}
