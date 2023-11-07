import { SelectableValue } from '@grafana/data';

import {
  QueryType,
} from './types';

export interface QueryTypeInfo extends SelectableValue<QueryType> {
  value: QueryType; // not optional
}

export const queryTypeInfos: QueryTypeInfo[] = [
  {
    label: 'Get metric history',
    value: QueryType.GetMetricHistory,
    description: `Gets the history of a metric.`,
  },
  {
    label: 'Get metric value',
    value: QueryType.GetMetricValue,
    description: `Gets a metrics current value.`,
  },
  {
    label: 'Get metric aggregate',
    value: QueryType.GetMetricAggregate,
    description: `Gets a metrics aggregate value.`,
  },
];

