import { Metric, MyQuery } from './types';

/**
 * convert a legacy query to a valid query definition
 * @param query query which might contain a legacy metric definition using deprecated metricId, metricName fields.
 */
export function convertQuery(query: MyQuery): MyQuery {
  return {
    ...query,
    metricName: undefined,
    metricId: undefined,
    metrics: convertMetrics(query),
  };
}

/**
 * converts query metrics to a uniform metric array
 * @param query query which might contain legacy metric definition using deprecated fields
 */
export function convertMetrics(query: MyQuery): Metric[] | undefined {
  if (query.metrics) {
    return query.metrics;
  }
  if (query.metricId) {
    return [{ metricId: query.metricId, metricName: query.metricName }];
  }
  return undefined;
}
