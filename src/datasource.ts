import {
  DataFrame,
  DataFrameView,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
  MetricFindValue,
  ScopedVars,
  SelectableValue,
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
} from './types';
import { from, lastValueFrom, Observable, concat } from 'rxjs';
import { map, mergeMap, toArray } from 'rxjs/operators';
import { getRequestLooper, MultiRequestTracker } from './requestLooper';
import { appendMatchingFrames } from './appendFrames';
import { convertMetrics, convertQuery } from './convert';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
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
    return metric.metricName || metric.metricId || '';
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

  // parses a string from a variable query into Dimensions
  // query has the format dimension1=value1;dimension2=value.... split this string into {dimension1: value1, dimension2: value2}
  parseDimensions(query: string): Dimensions {
    return query
      .split(';')
      .map((x) => x.split('='))
      .filter((x) => x.length === 2)
      .map((v) => ({
        id: '',
        key: v[0],
        value: v[1],
      }));
  }

  /**
   * Supports lists of metrics
   */
  async metricFindQuery(query: string, options?: any): Promise<MetricFindValue[]> {
    const dimensions = this.parseDimensions(query);

    const metrics = this.runListMetricsQuery(dimensions, '').pipe(map((x) => x.map((x) => ({ text: x.value || '' }))));

    return lastValueFrom(metrics);
  }

  /**
   * Supports template variables for metricId
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

    return {
      ...query2,
      metrics: metrics || [],
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

  runListMetricsQuery(dimensions: Dimensions, filter: string): Observable<Array<SelectableValue<string>>> {
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

  listMetrics(dimensions: Dimensions, filter: string): Observable<Array<SelectableValue<string>>> {
    const remoteMetrics = this.runListMetricsQuery(dimensions, filter).pipe(mergeMap((x) => x.flat()));

    const variables = from(
      getTemplateSrv()
        .getVariables()
        .map((x) => ({
          value: `$${x.name}`,
          label: `$${x.name}`,
        }))
    );

    return concat(remoteMetrics, variables).pipe(toArray());
  }
}

let counter = 1000;
