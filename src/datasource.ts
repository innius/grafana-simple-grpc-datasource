import {
  DataFrame,
  DataFrameView,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  ScopedVars,
  SelectableValue,
  MetricFindValue,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import {
  Dimension,
  Dimensions,
  isMetricQuery,
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
  OptionValue,
  VariableQuery,
  VariableQueryType,
} from './types';
import { lastValueFrom, Observable } from 'rxjs';
import { map } from 'rxjs/operators';
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
      process: (t: MultiRequestTracker, data: DataFrame[], isLast: boolean) => {
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
      onCancel: (tracker: MultiRequestTracker) => {},
    });
  }

  filterQuery(query: MyQuery): boolean {
    if (query.hide) {
      return false;
    }
    if (!query.queryType) {
      return false;
    }
    if (!isMetricQuery(query.queryType)) {
      return true;
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
  async metricFindQuery(query: VariableQuery, options?: any): Promise<MetricFindValue[]> {
    const q = query;

    if (q.queryType === VariableQueryType.dimensionValue) {
      if (!q.dimensionKey) {
        return [];
      }
      const values = await this.listDimensionsValues(q.dimensionKey, q.dimensionValueFilter || '', []);
      return values.map((x) => ({ text: x.value || '' }));
    }

    const metrics = this.listMetrics(q.dimensions, '').pipe(map((x) => x.map((x) => ({ text: x.value || '' }))));
    return lastValueFrom(metrics);
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

  async listDimensionKeys(filter: string, selected_dimensions: Dimensions): Promise<Array<SelectableValue<string>>> {
    const query: ListDimensionsQuery = {
      refId: 'listDimensionKeys',
      queryType: QueryType.ListDimensionKeys,
      selected_dimensions,
      filter: filter,
    };
    const dimKeys = this.runQuery(query).pipe(
      map((res) => {
        if (res.data.length) {
          const dimensions = new DataFrameView<SelectableValue<string>>(res.data[0]);
          return dimensions.toArray();
        }
        throw `no dimensions found ${res.error}`;
      })
    );
    return lastValueFrom(dimKeys);
  }

  async listDimensionsValues(
    key: string,
    filter: string,
    selected_dimensions: Dimensions
  ): Promise<Array<SelectableValue<string>>> {
    const query: ListDimensionValuesQuery = {
      refId: 'listDimensionsValues',
      queryType: QueryType.ListDimensionValues,
      dimensionKey: key,
      selected_dimensions,
      filter: filter,
    };

    const dimValues = this.runQuery(query).pipe(
      map((res) => {
        if (res.data.length) {
          const dimensionValues = new DataFrameView<SelectableValue<string>>(res.data[0]);
          return dimensionValues.toArray();
        }
        throw 'no dimension values found';
      })
    );
    return lastValueFrom(dimValues);
  }

  listMetrics(dimensions: Dimensions, filter: string): Observable<Array<SelectableValue<string>>> {
    const query: ListMetricsQuery = {
      refId: 'listMetrics',
      queryType: QueryType.ListMetrics,
      dimensions: dimensions,
      filter: filter,
    };

    return this.runQuery(query).pipe(
      map((res) => {
        if (res.data.length) {
          const metrics = new DataFrameView<SelectableValue<string>>(res.data[0]);
          return metrics.toArray();
        }
        throw 'no metrics found';
      })
    );
  }

  async getQueryOptions(qt: QueryType): Promise<QueryOptions> {
    return this.getResource<QueryOptions>('/options', { query_type: qt });
  }
}

function cloneQueryOptionsWithModifiedValues(
  queryOptions: { [key: string]: OptionValue },
  replace: (x: string) => string
) {
  const clonedOptions = queryOptions || {};

  for (const key in queryOptions) {
    if (queryOptions.hasOwnProperty(key)) {
      const {label, value} = queryOptions[key];
      clonedOptions[key] = { label: replace(label!), value: replace(value!) };
    }
  }

  return clonedOptions;
}

let counter = 1000;
