import { DataQuery, DataSourceJsonData } from '@grafana/schema';

export enum QueryType {
  GetMetricValue = 'GetMetricValue',
  GetMetricHistory = 'GetMetricHistory',
  GetMetricAggregate = 'GetMetricAggregate',
}

export interface Metric {
  metricId?: string;
}

// OptionValue is the selected value for a backend define query option
export interface QueryOptionValue {
  value?: string;
  label?: string;
}

// OptionValues are the query options which originate from the backend
// and are sent along with the query request
export type QueryOptions = { [key: string]: QueryOptionValue };

export interface StreamingConfig {
  // Maximum number of datapoints to keep in buffer
  maxBufferSize?: number;
  // Look-back period in seconds for initial dataset
  lookBackPeriod?: string;
  // Whether streaming is supported for this query
  streamingSupported?: boolean;
  // Error message if streaming is not supported
  errorMessage?: string;
}

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

  // the query is a streaming query - can be boolean or string (for variables)
  isStreaming?: boolean | string;

  // streaming configuration
  streamingConfig?: StreamingConfig;
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
  Enum = 'Enum',
  Boolean = 'Boolean',
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

export interface DimensionKeyDefinition {
  value?: string;
  label?: string;
  description?: string;
}

export interface DimensionValueDefinition {
  value?: string;
  label?: string;
  description?: string;
}

export interface MetricDefinition {
  value?: string;
  label?: string;
  description?: string;
}

export const defaultQuery: Partial<MyQuery> = {
  dimensions: [],
  queryType: QueryType.GetMetricAggregate,
  queryOptions: {},
  streamingConfig: {
    maxBufferSize: undefined,
    lookBackPeriod: undefined, // 5 minutes default
  },
};

export const defaultDataSourceOptions: Partial<MyDataSourceOptions> = {
  max_retries: 5,
  enableStreaming: false,
};

/**
 are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  endpoint?: string;
  apikey_authentication_enabled: boolean;

  // max. number of retries for all backend requests
  max_retries?: number;

  // enable streaming queries at datasource level
  enableStreaming?: boolean;
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

export interface ListDimensionsQuery {
  selected_dimensions: Dimensions;
  filter: string;
}

export interface ListDimensionValuesQuery {
  selected_dimensions: Dimensions;
  dimensionKey: string;
  filter: string;
}

export interface ListMetricsQuery {
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
