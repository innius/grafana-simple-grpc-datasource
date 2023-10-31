import { DataQuery, DataSourceJsonData } from '@grafana/schema';

export enum QueryType {
  ListDimensionKeys = 'ListDimensionKeys',
  ListDimensionValues = 'ListDimensionValues',
  ListMetrics = 'ListMetrics',
  GetMetricValue = 'GetMetricValue',
  GetMetricHistory = 'GetMetricHistory',
  GetMetricAggregate = 'GetMetricAggregate',
}

export function isMetricQuery(queryType: QueryType): boolean {
  return (
    queryType === QueryType.GetMetricValue ||
    queryType === QueryType.GetMetricHistory ||
    queryType === QueryType.GetMetricAggregate
  );
}

export interface Metric {
  metricId?: string;
}

// OptionValue is the selected value for a backend define query option  
export interface QueryOptionValue {
  value?: string
  label?: string
}

// OptionValues are the query options which originate from the backend 
// and are sent along with the query request
export type QueryOptions = { [key: string]: QueryOptionValue }

export interface MyQuery extends DataQuery {
  queryType: QueryType;
  dimensions?: Dimensions;
  metrics?: Metric[];

  /**
   * @deprecated use queryOptions instead
   */
  aggregateType?: string;
  displayName?: string;

  /**
   * @deprecated use metrics
   */
  metricId?: string;

  queryOptions?: QueryOptions;
}

export interface NextQuery extends MyQuery {
  /**
   * The next token should never be saved in the JSON model, however some queries
   * will require multiple pages in order to fulfil the requests
   */
  nextToken?: string;
}

export interface Metadata {
  nextToken?: string;
}

export interface EnumValue {
  label: string;
  description: string;
  id: string;
  default: boolean;
}

export enum OptionType {
  Enum = "Enum",
  Boolean = "Boolean"
}

export interface QueryOptionDefinition {
  label: string;
  id: string;
  description: string;
  type: OptionType;
  enumValues: EnumValue[];
  required: boolean;
}

export type QueryOptionDefinitions = QueryOptionDefinition[];

export interface Dimension {
  id: string;
  key: string;
  value: string;
}

export type Dimensions = Dimension[];

export const defaultQuery: Partial<MyQuery> = {
  dimensions: [],
  queryType: QueryType.GetMetricHistory,
  queryOptions: {},
};

export const defaultDataSourceOptions: Partial<MyDataSourceOptions> = {
  max_retries: 5,
};

/**
 are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  endpoint?: string;
  apikey_authentication_enabled: boolean;

  // max. number of retries for all backend requests
  max_retries?: number;
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
  selected_dimensions: Dimensions;
  filter: string;
}

export interface ListDimensionValuesQuery extends MyQuery {
  queryType: QueryType.ListDimensionValues;
  selected_dimensions: Dimensions;
  dimensionKey: string;
  filter: string;
}

export interface ListMetricsQuery extends MyQuery {
  queryType: QueryType.ListMetrics;
  dimensions: Dimensions;
  filter: string;
}

export enum VariableQueryType {
  metric = 'metric',
  dimensionValue = 'dimension Value',
}

export interface VariableQuery extends DataQuery {
  queryType: VariableQueryType;
  dimensionKey?: string;
  dimensions: Dimension[];
  dimensionValueFilter?: string;
}

