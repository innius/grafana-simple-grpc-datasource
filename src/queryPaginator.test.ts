import { DataQueryRequest, DataQueryResponse, DateTime, LoadingState } from '@grafana/data';

import { MyQueryPaginator } from 'queryPaginator';
import { QueryType, NextQuery, MyQuery } from 'types';
import { first, last } from 'rxjs/operators';

const dataQueryRequest: DataQueryRequest<MyQuery> = {
  app: 'panel-editor',
  requestId: 'Q112',
  timezone: 'browser',
  panelId: 2,
  dashboardUID: 'OPixSZySk',
  range: {
    from: new Date('2024-05-28T20:59:49.659Z') as unknown as DateTime,
    to: new Date('2024-05-28T21:29:49.659Z') as unknown as DateTime,
    raw: {
      from: 'now-30m',
      to: 'now',
    },
  },
  interval: '2s',
  intervalMs: 2000,
  targets: [
    {
      datasource: {
        type: 'innius-grpc-datasource',
        uid: 's0PWceLIz',
      },
      metrics: [{ metricId: 'foo' }],
      queryType: QueryType.GetMetricValue,
      refId: 'A',
    },
  ],
  maxDataPoints: 711,
  scopedVars: {
    __interval: {
      text: '2s',
      value: '2s',
    },
    __interval_ms: {
      text: '2000',
      value: 2000,
    },
  },
  startTime: 1716931789659,
  rangeRaw: {
    from: 'now-30m',
    to: 'now',
  },
};

// Response with MyQuery data
const dataQueryResponse: DataQueryResponse = {
  data: [
    {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: 'time',
          typeInfo: {
            frame: 'time.Time',
          },
          config: {},
          values: [1716931550000],
          entities: {},
        },
        {
          name: 'RotationsPerSecond',
          type: 'number',
          typeInfo: {
            frame: 'float64',
          },
          config: {
            unit: 'RPS',
          },
          values: [0.45253960150485795],
          entities: {},
        },
        {
          name: 'quality',
          type: 'string',
          typeInfo: {
            frame: 'string',
          },
          config: {},
          values: ['GOOD'],
          entities: {},
        },
      ],
      length: 1,
    },
  ],
  state: LoadingState.Done,
};

const dataQueryRequestPaginating: DataQueryRequest<NextQuery> = {
  app: 'panel-editor',
  requestId: 'Q112.2',
  timezone: 'browser',
  panelId: 2,
  dashboardUID: 'OPixSZySk',
  range: {
    from: new Date('2024-05-28T20:59:49.659Z') as unknown as DateTime,
    to: new Date('2024-05-28T21:29:49.659Z') as unknown as DateTime,
    raw: {
      from: 'now-30m',
      to: 'now',
    },
  },
  interval: '2s',
  intervalMs: 2000,
  targets: [
    {
      datasource: {
        type: 'innius-grpc-datasource',
        uid: 's0PWceLIz',
      },
      queryType: QueryType.GetMetricValue,
      metrics: [{ metricId: 'foo' }],
      refId: 'A',
      nextToken: 'mock-next-token-value',
    },
  ],
  maxDataPoints: 711,
  scopedVars: {
    __interval: {
      text: '2s',
      value: '2s',
    },
    __interval_ms: {
      text: '2000',
      value: 2000,
    },
  },
  startTime: 1716931789659,
  rangeRaw: {
    from: 'now-30m',
    to: 'now',
  },
};

const dataQueryResponsePaginating: DataQueryResponse = {
  data: [
    {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: 'time',
          typeInfo: {
            frame: 'time.Time',
          },
          config: {},
          values: [1716931549000],
          entities: {},
        },
        {
          name: 'RotationsPerSecond',
          type: 'number',
          typeInfo: {
            frame: 'float64',
          },
          config: {
            unit: 'RPS',
          },
          values: [1],
          entities: {},
        },
        {
          name: 'quality',
          type: 'string',
          typeInfo: {
            frame: 'string',
          },
          config: {},
          values: ['GOOD'],
          entities: {},
        },
      ],
      length: 1,
      meta: {
        custom: {
          nextToken: 'mock-next-token-value',
          resolution: 'RAW',
        },
      },
    },
  ],
  state: LoadingState.Done,
};

// Response with data combined from `dataQueryResponse` and `dataQueryResponsePaginating`
const dataQueryResponseCombined: DataQueryResponse = {
  data: [
    {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: 'time',
          typeInfo: {
            frame: 'time.Time',
          },
          config: {},
          values: [1716931549000, 1716931550000],
          entities: {},
        },
        {
          name: 'RotationsPerSecond',
          type: 'number',
          typeInfo: {
            frame: 'float64',
          },
          config: {
            unit: 'RPS',
          },
          values: [1, 0.45253960150485795],
          entities: {},
        },
        {
          name: 'quality',
          type: 'string',
          typeInfo: {
            frame: 'string',
          },
          config: {},
          values: ['GOOD', 'GOOD'],
          entities: {},
        },
      ],
      length: 2,
    },
  ],
  state: LoadingState.Done,
};

describe('MyQueryPaginator', () => {
  describe('toObservable()', () => {
    it('handles single page request', async () => {
      const request = dataQueryRequest;
      const queryFn = jest.fn().mockResolvedValue(dataQueryResponse);

      const queryObservable = new MyQueryPaginator({
        request,
        queryFn,
      }).toObservable();

      const firstResponse = queryObservable.pipe(first()).toPromise();
      expect(firstResponse).resolves.toMatchObject(dataQueryResponse);

      const lastResponse = queryObservable.pipe(last()).toPromise();
      expect(lastResponse).resolves.toMatchObject(dataQueryResponse);

      await lastResponse;
      expect(queryFn).toHaveBeenCalledTimes(1);
      expect(queryFn).toHaveBeenCalledWith(request);
    });

    it('handles more than 1 page request', async () => {
      const request = dataQueryRequest;
      const queryFn = jest
        .fn()
        .mockResolvedValueOnce(dataQueryResponsePaginating)
        .mockResolvedValueOnce(dataQueryResponse);

      const queryObservable = new MyQueryPaginator({
        request,
        queryFn,
      }).toObservable();

      const firstResponse = queryObservable.pipe(first()).toPromise();
      expect(firstResponse).resolves.toMatchObject({
        ...dataQueryResponsePaginating,
        state: LoadingState.Streaming,
      });

      const lastResponse = queryObservable.pipe(last()).toPromise();
      expect(lastResponse).resolves.toMatchObject(dataQueryResponseCombined);

      await lastResponse;
      expect(queryFn).toHaveBeenCalledTimes(2);
      expect(queryFn).toHaveBeenCalledWith(request);
      expect(queryFn).toHaveBeenCalledWith(dataQueryRequestPaginating);
    });

    it('handles error state response and terminate pagination', async () => {
      const request = dataQueryRequest;
      const queryFn = jest
        .fn()
        .mockResolvedValueOnce({
          ...dataQueryResponsePaginating,
          state: LoadingState.Error,
        })
        .mockResolvedValueOnce(dataQueryResponse);

      const queryObservable = new MyQueryPaginator({
        request,
        queryFn,
      }).toObservable();

      const firstResponse = queryObservable.pipe(first()).toPromise();
      expect(firstResponse).resolves.toMatchObject({
        ...dataQueryResponsePaginating,
        state: LoadingState.Error,
      });

      const lastResponse = queryObservable.pipe(last()).toPromise();
      expect(lastResponse).resolves.toMatchObject({
        ...dataQueryResponsePaginating,
        state: LoadingState.Error,
      });

      await lastResponse;
      expect(queryFn).toHaveBeenCalledTimes(1);
      expect(queryFn).toHaveBeenCalledWith(request);
    });
  });
});
