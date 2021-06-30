import {
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
  Dimensions,
  isMetricQuery,
  ListDimensionsQuery,
  ListDimensionValuesQuery,
  ListMetricsQuery,
  MyDataSourceOptions,
  MyQuery,
  QueryType,
} from './types';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  query(request: DataQueryRequest<MyQuery>): Observable<DataQueryResponse> {
    return super.query(request);
  }

  filterQuery(query: MyQuery): boolean {
    if (!query.queryType) {
      return false;
    }
    return !(isMetricQuery(query.queryType) && !query.metricId);
  }

  getQueryDisplayText(query: MyQuery): string {
    const dimensions = query.dimensions?.map(x => `${x.key}=${x.value}`).join(',');
    let text = query.metricId || '';
    if (!!dimensions) {
      text = `[${dimensions}] ${text}`;
    }
    return text;
  }

  /**
   * Supports lists of metrics
   */
  async metricFindQuery(query: string, options?: any): Promise<MetricFindValue[]> {
    // query has the format dimension1=value1;dimension2=value.... split this string into {dimension1: value1, dimension2: value2}
    let dimensions = query
      .split(';')
      .map(x => x.split('='))
      .filter(x => x.length === 2)
      .map(v => ({
        id: '',
        key: v[0],
        value: v[1],
      }));
    const res = await this.listMetrics(dimensions, '');
    return res.map(x => ({ text: x.value || '' }));
  }

  /**
   * Supports template variables for metricId
   */
  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars): MyQuery {
    const templateSrv = getTemplateSrv();
    return {
      ...query,
      metricId: templateSrv.replace(query.metricId || '', scopedVars),
    };
  }

  runQuery(query: MyQuery, maxDataPoints?: number): Observable<DataQueryResponse> {
    // @ts-ignore
    return this.query({ targets: [query], requestId: `iot.${counter++}`, maxDataPoints });
  }

  async listDimensionKeys(filter: string): Promise<Array<SelectableValue<string>>> {
    const query: ListDimensionsQuery = {
      refId: 'listDimensionKeys',
      queryType: QueryType.ListDimensionKeys,
      filter: filter,
    };
    return this.runQuery(query)
      .pipe(
        map(res => {
          if (res.data.length) {
            const dimensions = new DataFrameView<SelectableValue<string>>(res.data[0]);
            return dimensions.toArray();
          }
          throw `no dimensions found ${res.error}`;
        })
      )
      .toPromise();
    // return this.getResource('dimensions', {filter: filter})
  }

  async listDimensionsValues(key: string, filter: string): Promise<Array<SelectableValue<string>>> {
    const query: ListDimensionValuesQuery = {
      refId: 'listDimensionsValues',
      queryType: QueryType.ListDimensionValues,
      dimensionKey: key,
      filter: filter,
    };
    return this.runQuery(query)
      .pipe(
        map(res => {
          if (res.data.length) {
            const dimensionValues = new DataFrameView<SelectableValue<string>>(res.data[0]);
            return dimensionValues.toArray();
          }
          throw 'no dimension values found';
        })
      )
      .toPromise();
  }

  async listMetrics(dimensions: Dimensions, filter: string): Promise<Array<SelectableValue<string>>> {
    const query: ListMetricsQuery = {
      refId: 'listMetrics',
      queryType: QueryType.ListMetrics,
      dimensions: dimensions,
      filter: filter,
    };
    const remoteMetrics = await this.runQuery(query)
      .pipe(
        map(res => {
          if (res.data.length) {
            const metrics = new DataFrameView<SelectableValue<string>>(res.data[0]);
            return metrics.toArray();
          }
          throw 'no metrics found';
        })
      )
      .toPromise();

    const variables = getTemplateSrv()
      .getVariables()
      .map(x => ({
        value: `$${x.name}`,
        label: `$${x.name}`,
      })) as Array<SelectableValue<string>>;

    return remoteMetrics ? variables.concat(remoteMetrics) : variables;
  }
}

let counter = 1000;
