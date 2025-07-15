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
  StreamingConfig,
} from './types';
import { convertMetrics, convertQuery } from './convert';
import { DatasourceVariableSupport } from './variables';
import { Observable, merge } from 'rxjs';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  enableStreaming: boolean;

  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
    this.enableStreaming = instanceSettings.jsonData.enableStreaming || false;
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

  // Maximum allowed channel length for streaming paths
  private static readonly MAX_CHANNEL_LENGTH = 100;

  /**
   * Sanitizes a string to be used in streaming path by removing/replacing invalid characters
   * Only allows letters, numbers and forward slashes
   */
  private sanitizePathComponent(input: string): string {
    // Replace invalid characters with underscores, allow only alphanumeric and forward slashes (no spaces)
    return input.replace(/[^a-zA-Z0-9\/]/g, '_');
  }

  /**
   * Creates a deterministic hash of the input string
   * Uses djb2 hash algorithm for good distribution and collision resistance
   */
  private createHash(input: string): string {
    let hash = 5381;
    for (let i = 0; i < input.length; i++) {
      hash = (hash << 5) + hash + input.charCodeAt(i);
    }
    // Convert to positive number and use base36 for compact representation
    return Math.abs(hash >>> 0).toString(36);
  }

  /**
   * Creates a compact string representation of all query components
   */
  private serializeQueryComponents(query: MyQuery): string {
    const parts: string[] = [];

    // Add query type
    if (query.queryType) {
      parts.push(`qt:${query.queryType}`);
    }

    // Add first metric if available
    if (query.metrics && query.metrics.length > 0 && query.metrics[0].metricId) {
      parts.push(`m:${query.metrics[0].metricId}`);
    }

    // Add dimensions if available
    if (query.dimensions && query.dimensions.length > 0) {
      const dims = query.dimensions.map((dim) => `${dim.key || ''}=${dim.value || ''}`).join(',');
      parts.push(`d:${dims}`);
    }

    // Add query options if available
    if (query.queryOptions && Object.keys(query.queryOptions).length > 0) {
      const opts = Object.entries(query.queryOptions)
        .filter(([_, optionValue]) => optionValue.value)
        .map(([key, optionValue]) => `${key}=${optionValue.value || ''}`)
        .join(',');
      if (opts) {
        parts.push(`o:${opts}`);
      }
    }

    // Add streaming configuration
    if (query.streamingConfig) {
      const streamingParts: string[] = [];
      if (query.streamingConfig.maxBufferSize) {
        streamingParts.push(`b${query.streamingConfig.maxBufferSize}`);
      }
      if (query.streamingConfig.lookBackPeriod) {
        streamingParts.push(`l${query.streamingConfig.lookBackPeriod}`);
      }
      if (streamingParts.length > 0) {
        parts.push(`s:${streamingParts.join(',')}`);
      }
    }

    return parts.join('|');
  }

  /**
   * Creates a properly formatted path for streaming queries with length constraints
   * Uses hashing to ensure paths stay under the 100-character limit while maintaining uniqueness
   * Format: refId/hash or hash (if refId is too long)
   */
  private createStreamingPath(query: MyQuery): string {
    const refId = this.sanitizePathComponent(query.refId || 'default');
    const queryComponents = this.serializeQueryComponents(query);

    // If no additional components, try to use just refId
    if (!queryComponents) {
      return refId.length <= DataSource.MAX_CHANNEL_LENGTH ? refId : this.createHash(refId);
    }

    // Create hash of all query components for uniqueness
    const queryHash = this.createHash(queryComponents);

    // Try to include refId if there's space
    const pathWithRefId = `${refId}/${queryHash}`;

    if (pathWithRefId.length <= DataSource.MAX_CHANNEL_LENGTH) {
      return pathWithRefId;
    }

    // If refId + hash is too long, try with truncated refId
    const maxRefIdLength = DataSource.MAX_CHANNEL_LENGTH - queryHash.length - 1; // -1 for slash
    if (maxRefIdLength > 0) {
      const truncatedRefId = refId.substring(0, maxRefIdLength);
      return `${truncatedRefId}/${queryHash}`;
    }

    // If even truncated refId doesn't fit, hash everything together
    const fullHash = this.createHash(`${refId}|${queryComponents}`);
    return fullHash.length <= DataSource.MAX_CHANNEL_LENGTH
      ? fullHash
      : fullHash.substring(0, DataSource.MAX_CHANNEL_LENGTH);
  }

  query(options: DataQueryRequest<MyQuery>): Observable<DataQueryResponse> {
    const streams: Array<Observable<DataQueryResponse>> = [];
    const backendQueries: MyQuery[] = [];
    const streamingVerifications: Promise<void>[] = [];

    for (let target of options.targets) {
      // Apply template variables first to resolve streaming state
      const resolvedTarget = this.applyTemplateVariables(target, options.scopedVars);

      if (this.isStreamingEnabled(resolvedTarget)) {
        // Create a verification promise for this query
        const verificationPromise = this.verifyStreamingSupport(resolvedTarget).then((config) => {
          if (config.streamingSupported) {
            // Update the streaming config with the verified configuration
            const updatedTarget = {
              ...resolvedTarget,
              streamingConfig: {
                ...resolvedTarget.streamingConfig,
                ...config,
              },
            };
            streams.push(this.runGrafanaLiveQuery(updatedTarget, options));
          } else {
            // If streaming is not supported, throw an error to show in Grafana UI
            throw new Error(
              `Streaming not supported for query ${resolvedTarget.refId}: ${
                config.errorMessage || 'No reason provided'
              }`
            );
          }
        });
        streamingVerifications.push(verificationPromise);
      } else {
        backendQueries.push(resolvedTarget);
      }
    }

    // Return an observable that waits for all streaming verifications to complete
    return new Observable<DataQueryResponse>((subscriber) => {
      let subscription: any = null;

      Promise.all(streamingVerifications)
        .then(() => {
          const observables: Array<Observable<DataQueryResponse>> = [];

          // Add all verified streaming queries
          if (streams.length) {
            observables.push(merge(...streams));
          }

          // Add all non-streaming queries
          if (backendQueries.length) {
            const backendOpts = {
              ...options,
              targets: backendQueries,
            };
            observables.push(super.query(backendOpts));
          }

          // If no queries at all, return empty result
          if (observables.length === 0) {
            subscriber.next({ data: [] });
            subscriber.complete();
            return;
          }

          // Merge all observables and forward their events
          subscription = merge(...observables).subscribe({
            next: (response) => subscriber.next(response),
            error: (error) => subscriber.error(error),
            complete: () => subscriber.complete(),
          });
        })
        .catch((error) => {
          subscriber.error(error);
        });

      // Return cleanup function that will be called when the outer Observable is unsubscribed
      return () => {
        if (subscription) {
          subscription.unsubscribe();
        }
      };
    });
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

  runGrafanaLiveQuery(target: MyQuery, req: DataQueryRequest<MyQuery>): Observable<DataQueryResponse> {
    const path = this.createStreamingPath(target);
    const streamingConfig = target.streamingConfig || { maxBufferSize: 3600, lookBackPeriod: '1h' };

    // Check if streaming is explicitly not supported
    if (streamingConfig.streamingSupported === false) {
      return new Observable<DataQueryResponse>((subscriber) => {
        subscriber.error(new Error(`Streaming not supported: ${streamingConfig.errorMessage || 'No reason provided'}`));
      });
    }

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
          lookBackPeriod: streamingConfig.lookBackPeriod,
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

  /**
   * Verifies if streaming is supported for a given query
   * Returns the streaming configuration with streamingSupported flag
   * Throws errors that should be visible to Grafana users
   */
  async verifyStreamingSupport(query: MyQuery): Promise<StreamingConfig> {
    const response = await this.postResource<StreamingConfig>('streaming/verify', { ...query });
    return {
      // maxBufferSize: response.maxBufferSize || 3600,
      // lookBackPeriod: response.lookBackPeriod || '1h',
      streamingSupported: response.streamingSupported !== undefined ? response.streamingSupported : true,
      errorMessage: response.errorMessage,
    };
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
