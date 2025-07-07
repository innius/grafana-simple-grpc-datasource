import {
  DataSourceInstanceSettings,
  ScopedVars,
  MetricFindValue,
  DataQueryRequest,
  DataQueryResponse,
  LiveChannelScope,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv, getGrafanaLiveSrv } from '@grafana/runtime';
import {
  Dimension,
  Dimensions,
  ListDimensionsQuery,
  ListDimensionValuesQuery,
  ListMetricsQuery,
  Metric,
  MyDataSourceOptions,
  MyQuery,
  QueryType,
  QueryOptions,
  QueryOptionDefinitions,
  QueryOptionValue,
  VariableQuery,
  VariableQueryType,
  DimensionKeyDefinition,
  DimensionValueDefinition,
  MetricDefinition,
} from './types';
import { convertMetrics, convertQuery } from './convert';
import { DatasourceVariableSupport } from './variables';
import { Observable, of, merge } from 'rxjs';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
    this.variables = new DatasourceVariableSupport(this);
  }

  filterQuery(query: MyQuery): boolean {
    if (query.hide) {
      return false;
    }
    if (!query.queryType) {
      return false;
    }
    const metrics = convertMetrics(query);
    return metrics !== undefined && metrics.length > 0;
  }

  formatMetric(metric: Metric): string {
    return metric.metricId || '';
  }

  formatDimension(dim: Dimension): string {
    return `${dim.key}=${dim.value}`;
  }

  getQueryDisplayText(query: MyQuery): string {
    let displayText = `[${query.queryType || 'Unknown'}]`;
    
    if (query.dimensions && query.dimensions.length > 0) {
      displayText += '[' + query.dimensions.map(this.formatDimension).join(',') + ']';
    }

    if (query.metrics && query.metrics?.length > 0) {
      displayText += ' ' + query.metrics.map(this.formatMetric).join('&');
    }

    // Add query options if available
    if (query.queryOptions && Object.keys(query.queryOptions).length > 0) {
      const optionsStr = Object.entries(query.queryOptions)
        .filter(([_, optionValue]) => optionValue.value)
        .map(([key, optionValue]) => `${key}=${optionValue.value}`)
        .join(',');
      if (optionsStr) {
        displayText += ` {${optionsStr}}`;
      }
    }

    return displayText || query.refId;
  }

  /**
   * Sanitizes a string to be used in streaming path by removing/replacing invalid characters
   * Only allows letters, numbers and forward slashes
   */
  private sanitizePathComponent(input: string): string {
    // Replace invalid characters with underscores, allow only alphanumeric and forward slashes (no spaces)
    return input.replace(/[^a-zA-Z0-9\/]/g, '_');
  }

  /**
   * Creates a hash-based path component for very long paths to avoid URL length issues
   */
  private createHashedPathComponent(input: string): string {
    // Simple hash function for creating shorter unique identifiers
    let hash = 0;
    for (let i = 0; i < input.length; i++) {
      const char = input.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // Convert to 32-bit integer
    }
    return Math.abs(hash).toString(36);
  }

  /**
   * Creates a properly formatted path for streaming queries
   * Format: /refId/queryType/metricId/dimensions/queryOptions
   * Uses hashing for very long paths to avoid URL length issues
   */
  private createStreamingPath(query: MyQuery): string {
    let path = `${this.sanitizePathComponent(query.refId)}`;

    // Add query type
    if (query.queryType) {
      path += `/${this.sanitizePathComponent(query.queryType)}`;
    }

    // Add first metric if available
    if (query.metrics && query.metrics.length > 0 && query.metrics[0].metricId) {
      const metricId = this.sanitizePathComponent(query.metrics[0].metricId);
      path += `/${metricId}`;
    }

    // Add dimensions if available
    if (query.dimensions && query.dimensions.length > 0) {
      const dimensionsStr = query.dimensions
        .map((dim) => `${this.sanitizePathComponent(dim.key || '')}/${this.sanitizePathComponent(dim.value || '')}`)
        .join('/');
      path += `/${dimensionsStr}`;
    }

    // Add query options if available
    if (query.queryOptions && Object.keys(query.queryOptions).length > 0) {
      const optionsStr = Object.entries(query.queryOptions)
        .filter(([_, optionValue]) => optionValue.value) // Only include options with values
        .map(([key, optionValue]) => `${this.sanitizePathComponent(key)}/${this.sanitizePathComponent(optionValue.value || '')}`)
        .join('/');
      if (optionsStr) {
        path += `/${optionsStr}`;
      }
    }

    // Add streaming configuration as part of the path for uniqueness
    if (query.streamingConfig) {
      const streamingStr = [
        query.streamingConfig.maxBufferSize ? `buffer/${query.streamingConfig.maxBufferSize}` : '',
        query.streamingConfig.lookBackPeriod ? `lookback/${query.streamingConfig.lookBackPeriod}` : ''
      ].filter(Boolean).join('/');
      
      if (streamingStr) {
        path += `/${streamingStr}`;
      }
    }

    // If the path is too long (>200 characters), use a hash-based approach
    if (path.length > 200) {
      const fullQueryString = JSON.stringify({
        refId: query.refId,
        queryType: query.queryType,
        metrics: query.metrics,
        dimensions: query.dimensions,
        queryOptions: query.queryOptions,
        streamingConfig: query.streamingConfig
      });
      
      const hashedPath = this.createHashedPathComponent(fullQueryString);
      path = `${this.sanitizePathComponent(query.refId)}/hashed/${hashedPath}`;
    }

    return path;
  }

  query(options: DataQueryRequest<MyQuery>): Observable<DataQueryResponse> {
    const streams: Array<Observable<DataQueryResponse>> = [];
    const backendQueries: MyQuery[] = [];
    for (let target of options.targets) {
      // Apply template variables first to resolve streaming state
      const resolvedTarget = this.applyTemplateVariables(target, options.scopedVars);
      
      if (this.isStreamingEnabled(resolvedTarget)) {
        streams.push(this.runGrafanaLiveQuery(resolvedTarget, options));
      } else {
        backendQueries.push(resolvedTarget);
      }
    }

    if (backendQueries.length) {
      const backendOpts = {
        ...options,
        targets: backendQueries,
      };
      streams.push(super.query(backendOpts));
    }
    if (streams.length === 0) {
      return of({ data: [] });
    }
    return merge(...streams);
  }

  /**
   * Helper method to determine if streaming is enabled for a query
   */
  private isStreamingEnabled(query: MyQuery): boolean {
    if (typeof query.isStreaming === 'boolean') return query.isStreaming;
    if (typeof query.isStreaming === 'string') {
      const lowerValue = query.isStreaming.toLowerCase().trim();
      return lowerValue === 'true' || lowerValue === '1' || lowerValue === 'yes';
    }
    return false;
  }

  /**
   * Helper method to get debug information about streaming path creation
   * Useful for troubleshooting path generation
   */
  getStreamingPathDebugInfo(query: MyQuery): { path: string; components: any; isHashed: boolean } {
    const components = {
      refId: query.refId,
      queryType: query.queryType,
      firstMetric: query.metrics && query.metrics.length > 0 ? query.metrics[0].metricId : null,
      dimensionsCount: query.dimensions ? query.dimensions.length : 0,
      queryOptionsCount: query.queryOptions ? Object.keys(query.queryOptions).length : 0,
      streamingConfig: query.streamingConfig
    };

    const path = this.createStreamingPath(query);
    const isHashed = path.includes('/hashed/');

    return { path, components, isHashed };
  }

  runGrafanaLiveQuery(target: MyQuery, req: DataQueryRequest<MyQuery>): Observable<DataQueryResponse> {
    const path = this.createStreamingPath(target);
    const streamingConfig = target.streamingConfig || { maxBufferSize: 3600, lookBackPeriod: 300 };
    
    return getGrafanaLiveSrv().getDataStream({
      buffer: {
        maxLength: streamingConfig.maxBufferSize || 3600,
      },
      addr: {
        scope: LiveChannelScope.DataSource,
        namespace: this.uid,
        path: path,
        data: {
          range: req.range,
          intervalMs: req.intervalMs,
          maxDataPoints: req.maxDataPoints,
          lookBackPeriod: streamingConfig.lookBackPeriod || 300,
          ...target,
        },
      },
    });
  }
  /**
   * Supports lists of metrics
   */
  async metricFindQuery(query: VariableQuery, _?: any): Promise<MetricFindValue[]> {
    const q = query;

    if (q.queryType === VariableQueryType.dimensionValue) {
      if (!q.dimensionKey) {
        return [];
      }
      const values = await this.listDimensionsValues(q.dimensionKey, q.dimensionValueFilter || '', []);
      return values.map((x) => ({ text: x.value || '' }));
    }

    const metrics = await this.listMetrics(q.dimensions, '');

    return metrics.map((x) => ({ text: x.value || '' }));
  }

  /**
   * Supports template variables for metricId
   * one metric var may can be expanded into multiple metric
   * for example: [*] -> becomes ["a","b","c"]
   */
  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars): MyQuery {
    const templateSrv = getTemplateSrv();

    const query2 = convertQuery(query);
    const metrics = query2.metrics
      ?.flatMap<string[]>((metric) => {
        const replaced = templateSrv.replace(metric.metricId, scopedVars, 'json');
        try {
          return JSON.parse(replaced);
        } catch (e) {
          return [replaced];
        }
      })
      .flat()
      .map((x) => ({ metricId: x }));

    const dimensions = query2.dimensions?.map((x) => ({
      ...x,
      value: templateSrv.replace(x.value, scopedVars),
    }));

    const { queryOptions } = query2;

    // Handle template variables in isStreaming field
    let resolvedIsStreaming = query2.isStreaming;
    if (typeof query2.isStreaming === 'string') {
      resolvedIsStreaming = templateSrv.replace(query2.isStreaming, scopedVars);
    }

    return {
      ...query2,
      dimensions: dimensions,
      metrics: metrics || [],
      queryOptions: cloneQueryOptionsWithModifiedValues(queryOptions!, (x) => templateSrv.replace(x, scopedVars)),
      streamingConfig: query2.streamingConfig, // Preserve streaming configuration
      isStreaming: resolvedIsStreaming,
    };
  }

  async listDimensionKeys(filter: string, selected_dimensions: Dimensions): Promise<DimensionKeyDefinition[]> {
    const query: ListDimensionsQuery = {
      selected_dimensions,
      filter: filter,
    };
    return this.postResource<DimensionKeyDefinition[]>('dimensions', query);
  }

  async listDimensionsValues(
    key: string,
    filter: string,
    selected_dimensions: Dimensions
  ): Promise<DimensionValueDefinition[]> {
    if (key === '') {
      return Promise.resolve([]);
    }
    const query: ListDimensionValuesQuery = {
      dimensionKey: key,
      selected_dimensions,
      filter: filter,
    };
    return this.postResource<DimensionValueDefinition[]>('dimensions/values', query);
  }

  async listMetrics(dimensions: Dimensions, filter: string): Promise<MetricDefinition[]> {
    // Checking if 'dimensions' is undefined
    if (!dimensions || !dimensions.length) {
      return Promise.resolve([]);
    }
    // Filtering out empty dimensions (where Key is empty or undefined)
    const validDimensions = dimensions.filter((dim) => dim.value && dim.value !== '');

    // Checking if there are no valid dimensions, returning an empty array
    if (validDimensions.length === 0) {
      return Promise.resolve([]);
    }

    // Accessing the template service
    const templateSrv = getTemplateSrv();

    // Transforming dimensions by replacing their values
    const query: ListMetricsQuery = {
      dimensions: validDimensions.map((dim) => ({
        ...dim,
        value: templateSrv.replace(dim.value, {}),
      })),
      filter: filter,
    };

    // Making a POST request to 'metrics' endpoint with the constructed query
    return this.postResource('metrics', query);
  }

  async getQueryOptionDefinitions(qt: QueryType, opts: QueryOptions): Promise<QueryOptionDefinitions> {
    let selected: {
      [key: string]: string | undefined;
    } = {};
    Object.keys(opts).forEach((k) => {
      selected[k] = opts[k].value;
    });
    const query = {
      selected_options: selected,
      query_type: qt,
    };
    return this.postResource<QueryOptionDefinitions>('options', query);
  }
}

function cloneQueryOptionsWithModifiedValues(
  queryOptionValues: { [key: string]: QueryOptionValue },
  replace: (x: string) => string
) {
  const clonedOptions = queryOptionValues || {};

  for (const key in queryOptionValues) {
    if (queryOptionValues.hasOwnProperty(key)) {
      const { label, value } = queryOptionValues[key];
      clonedOptions[key] = { label: replace(label!), value: replace(value!) };
    }
  }

  return clonedOptions;
}
