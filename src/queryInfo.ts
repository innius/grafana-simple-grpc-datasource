import { SelectableValue } from '@grafana/data';

import { GetMetricHistoryQuery, GetMetricValueQuery, MyQuery, QueryType } from './types';

export interface QueryTypeInfo extends SelectableValue<QueryType> {
  value: QueryType; // not optional
  defaultQuery: Partial<MyQuery>;
}

export const queryTypeInfos: QueryTypeInfo[] = [
  {
    label: 'Get metric history',
    value: QueryType.GetMetricHistory,
    description: `Gets the history of a metric.`,
    defaultQuery: {} as GetMetricHistoryQuery,
  },
  {
    label: 'Get metric value',
    value: QueryType.GetMetricValue,
    description: `Gets a metrics current value.`,
    defaultQuery: {} as GetMetricValueQuery,
  },
];

export function changeQueryType(q: MyQuery, info: QueryTypeInfo): MyQuery {
  if (q.queryType === info.value) {
    return q; // no change;
  }
  return {
    ...info.defaultQuery,
    ...q,
    queryType: info.value,
  };
}
