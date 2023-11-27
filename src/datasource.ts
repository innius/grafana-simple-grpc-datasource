import {
  DataFrame,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  ScopedVars,
  MetricFindValue,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import {
  Dimension,
  Dimensions,
  ListDimensionsQuery,
  ListDimensionValuesQuery,
  ListMetricsQuery,
  Metadata,
  Metric,
  MyDataSourceOptions,
  MyQuery,
  NextQuery,
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
import { Observable } from 'rxjs';
import { getRequestLooper, MultiRequestTracker } from './requestLooper';
import { appendMatchingFrames } from './appendFrames';
import { convertMetrics, convertQuery } from './convert';
import { DatasourceVariableSupport } from './variables';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
    this.variables = new DatasourceVariableSupport(this);
  }
  query(request: DataQueryRequest<MyQuery>): Observable<DataQueryResponse> {
    return getRequestLooper(request, {
      // Check for a "nextToken" in the response
      getNextQueries: (rsp: DataQueryResponse) => {
        if (rsp.data?.length) {
          const next: NextQuery[] = [];
          for (const frame of rsp.data as DataFrame[]) {
            const meta = frame.meta?.custom as Metadata;
            if (meta && meta.nextToken) {
              const query = request.targets.find((t) => t.refId === frame.refId);
              if (query) {
                next.push({
                  ...query,
                  nextToken: meta.nextToken,
                });
              }
            }
          }
          if (next.length) {
            return next;
          }
        }
        return undefined;
      },
      /**
       * The original request
       */
      query: (request: DataQueryRequest<MyQuery>) => {
        return super.query(request);
      },

      /**
       * Process the results
       */
      process: (t: MultiRequestTracker, data: DataFrame[], _: boolean) => {
        if (t.data) {
          // append rows to fields with the same structure
          t.data = appendMatchingFrames(t.data, data);
        } else {
          t.data = data; // hang on to the results from the last query
        }
        return t.data;
      },

      /**
       * Callback that gets executed when unsubscribed
       */
      onCancel: (_: MultiRequestTracker) => { },
    });
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
    let displayText = '[' + query.dimensions?.map(this.formatDimension).join(',') + ']';

    if (query.metrics && query.metrics?.length > 0) {
      displayText += ' ' + query.metrics.map(this.formatMetric).join('&');
    }
    return displayText || query.refId;
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

    const metrics = await this.listMetrics(q.dimensions, '')

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

    return {
      ...query2,
      dimensions: dimensions,
      metrics: metrics || [],
      queryOptions: cloneQueryOptionsWithModifiedValues(queryOptions!, (x) => templateSrv.replace(x, scopedVars)),
    };
  }

  runQuery(query: MyQuery, maxDataPoints?: number): Observable<DataQueryResponse> {
    return this.query({
      targets: [query],
      requestId: `iot.${counter++}`,
      maxDataPoints,
    } as DataQueryRequest<MyQuery>);
  }

  async listDimensionKeys(filter: string, selected_dimensions: Dimensions): Promise<DimensionKeyDefinition[]> {
    const query: ListDimensionsQuery = {
      selected_dimensions,
      filter: filter,
    };
    return this.postResource<DimensionKeyDefinition[]>('dimensions', query);
  }

  async listDimensionsValues(key: string, filter: string, selected_dimensions: Dimensions): Promise<DimensionValueDefinition[]> {
    if (key === "") {
      return Promise.resolve([])
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
    const validDimensions = dimensions.filter(dim => dim.value && dim.value !== '');

    // Checking if there are no valid dimensions, returning an empty array
    if (validDimensions.length === 0) {
      return Promise.resolve([]);
    }

    // Accessing the template service
    const templateSrv = getTemplateSrv();

    // Transforming dimensions by replacing their values
    const query: ListMetricsQuery = {
      dimensions: validDimensions.map(dim => ({
        ...dim,
        value: templateSrv.replace(dim.value, {}),
      })),
      filter: filter,
    };

    // Making a POST request to 'metrics' endpoint with the constructed query
    return this.postResource('metrics', query);
  }

  async getQueryOptionDefinitions(qt: QueryType, opts: QueryOptions): Promise<QueryOptionDefinitions> {
    let selected: { [key: string]: string | undefined } = {}
    Object.keys(opts).forEach((k) => {
      selected[k] = opts[k].value
    })
    const query = {
      selected_options: selected,
      query_type: qt,
    }
    return this.postResource<QueryOptionDefinitions>('options', query)
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

let counter = 1000;
