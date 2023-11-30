import { Metric, MyQuery } from './types';

/**
 * convert a legacy query to a valid query definition
 * @param query query which might contain a legacy metric definition using deprecated metricId, metricName fields.
 */
export function convertQuery(query: MyQuery): MyQuery {
  let options = query.queryOptions;
  // convert deprecated aggregateType to query options
  if (!options && query.aggregateType) {
    const { aggregateType } = query;
    let aggregateTypeEnumValue = 0;
    switch (aggregateType.toUpperCase()) {
      case 'AVERAGE':
        aggregateTypeEnumValue = 0;
        break;
      case 'MAX':
        aggregateTypeEnumValue = 1;
        break;
      case 'MIN':
        aggregateTypeEnumValue = 2;
        break;
      case 'COUNT':
        aggregateTypeEnumValue = 3;
        break;
    }
    options = {
      "0": { value: aggregateTypeEnumValue.toString(), label: aggregateType.toLowerCase() },
    };
  }
  return {
    ...query,
    metricId: undefined,
    aggregateType: undefined,
    queryOptions: options,
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

  if ('metricId' in query) {
    console.warn('The "metricId" field is deprecated. Please use "metrics" instead.');
    return query.metricId ? [{ metricId: query.metricId }] : undefined;
  }
  return undefined;
}
