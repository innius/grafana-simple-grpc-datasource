import { FieldType, dateTime, LoadingState } from '@grafana/data';
import { RelativeRangeCache } from 'RelativeRangeRequestCache/RelativeRangeCache';
import { QueryType } from 'types';
import { generateRequestCacheId } from './cacheIdUtils';

describe('RelativeRangeCache', () => {
  const requestId = 'mock-request-id';
  const range = {
    from: dateTime('2024-05-28T00:00:00Z'),
    to: dateTime('2024-05-28T01:00:00Z'),
    raw: {
      from: 'now-1h',
      to: 'now',
    },
  };

  const request = {
    requestId,
    interval: '5s',
    intervalMs: 5000,
    range,
    scopedVars: {},
    targets: [
      {
        refId: 'A',
        queryType: QueryType.GetMetricHistory,
      },
      {
        refId: 'B',
        queryType: QueryType.GetMetricValue,
      },
    ],
    timezone: 'browser',
    app: 'dashboard',
    startTime: 1716858000000,
  };

  const requestDisabledCache = {
    ...request,
    targets: [
      {
        ...request.targets[0],
        clientCache: false,
      },
    ],
  };

  describe('get()', () => {
    it('returns undefined when any query with client cache disabled', () => {
      const cachedQueryInfo = [
        {
          query: {
            queryType: QueryType.GetMetricHistory,
            refId: 'A',
          },
          dataFrame: {
            name: 'Demo Turbine Asset 1',
            refId: 'A',
            fields: [
              {
                name: 'time',
                type: FieldType.time,
                config: {},
                values: [],
              },
            ],
            length: 0,
          },
        },
      ];
      const cacheData = {
        [generateRequestCacheId(requestDisabledCache)]: {
          queries: cachedQueryInfo,
          range,
        },
      };
      const cache = new RelativeRangeCache(new Map(Object.entries(cacheData)));

      expect(cache.get(requestDisabledCache)).toBeUndefined();
    });

    it('returns undefined when there is no cached response', () => {
      const cache = new RelativeRangeCache();

      expect(cache.get(request)).toBeUndefined();
    });

    it('returns starting cached data and time series query to fetch', () => {
      const cachedQueryInfo = [
        {
          query: {
            queryType: QueryType.GetMetricHistory,
            refId: 'A',
          },
          dataFrame: {
            name: 'Demo Turbine Asset 1',
            refId: 'A',
            fields: [
              {
                name: 'time',
                type: FieldType.time,
                config: {},
                values: [
                  1716854400000, // 2024-05-28T00:00:00Z
                  1716854400001, // 2024-05-28T00:15:00Z + 1ms
                  1716855300000, // 2024-05-28T00:15:00Z
                  1716855300001, // 2024-05-28T00:15:00Z + 1ms
                  1716857100000, // 2024-05-28T00:45:00Z
                  1716857100001, // 2024-05-28T00:45:00Z + 1ms
                ],
              },
              {
                name: 'RotationsPerSecond',
                type: FieldType.number,
                config: {
                  unit: 'RPS',
                },
                values: [0, 1, 2, 3, 4, 5],
              },
            ],
            length: 6,
          },
        },

        {
          query: {
            queryType: QueryType.GetMetricValue,
            refId: 'B',
          },
          dataFrame: {
            name: 'child',
            refId: 'B',
            fields: [
              {
                name: 'name',
                type: FieldType.string,
                config: {},
                values: ['child'],
              },
            ],
            length: 1,
          },
        },
      ];
      const expectedDataFrames = [
        {
          name: 'Demo Turbine Asset 1',
          refId: 'A',
          fields: [
            {
              name: 'time',
              type: FieldType.time,
              config: {},
              values: [
                1716854400001, // 2024-05-28T00:15:00Z + 1ms
                1716855300000, // 2024-05-28T00:15:00Z
                1716855300001, // 2024-05-28T00:15:00Z + 1ms
                1716857100000, // 2024-05-28T00:45:00Z
              ],
            },
            {
              name: 'RotationsPerSecond',
              type: FieldType.number,
              config: {
                unit: 'RPS',
              },
              values: [1, 2, 3, 4],
            },
          ],
          length: 4,
        },
        {
          name: 'child',
          refId: 'B',
          fields: [
            {
              name: 'name',
              type: FieldType.string,
              config: {},
              values: ['child'],
            },
          ],
          length: 1,
        },
      ];

      const cacheData = {
        [generateRequestCacheId(request)]: {
          queries: cachedQueryInfo,
          range,
        },
      };
      const cache = new RelativeRangeCache(new Map(Object.entries(cacheData)));

      const cacheResult = cache.get(request);

      expect(cacheResult).toBeDefined();
      expect(cacheResult?.cachedResponse).toEqual({
        start: {
          data: expectedDataFrames,
          key: requestId,
          state: LoadingState.Streaming,
        },
        end: {
          data: [],
          key: requestId,
          state: LoadingState.Streaming,
        },
      });
      expect(cacheResult?.refreshingRequest).toEqual({
        ...request,
        range: {
          from: dateTime(1716857100000), // '2024-05-28T01:45:00Z'
          to: dateTime('2024-05-28T01:00:00Z'),
          raw: {
            from: 'now-1h',
            to: 'now',
          },
        },
      });
    });
    describe('set()', () => {
      const cachedQueryInfo = [
        {
          query: {
            queryType: QueryType.GetMetricHistory,
            refId: 'A',
          },
          dataFrame: {
            name: 'Demo Turbine Asset 1',
            refId: 'A',
            fields: [
              {
                name: 'time',
                type: FieldType.time,
                config: {},
                values: [
                  1716854400001, // 2024-05-28T00:15:00Z + 1ms
                  1716855300000, // 2024-05-28T00:15:00Z
                  1716855300001, // 2024-05-28T00:15:00Z + 1ms
                  1716857100000, // 2024-05-28T00:45:00Z
                ],
              },
              {
                name: 'RotationsPerSecond',
                type: FieldType.number,
                config: {
                  unit: 'RPS',
                },
                values: [1, 2, 3, 4],
              },
            ],
            length: 4,
          },
        },
        {
          query: {
            queryType: QueryType.GetMetricValue,
            refId: 'B',
          },
          dataFrame: {
            name: 'child',
            refId: 'B',
            fields: [
              {
                name: 'name',
                type: FieldType.string,
                config: {},
                values: ['child'],
              },
            ],
            length: 1,
          },
        },
      ];
      const expectedDataFrames = [
        {
          name: 'Demo Turbine Asset 1',
          refId: 'A',
          fields: [
            {
              name: 'time',
              type: FieldType.time,
              config: {},
              values: [
                1716854400001, // 2024-05-28T00:15:00Z + 1ms
                1716855300000, // 2024-05-28T00:15:00Z
                1716855300001, // 2024-05-28T00:15:00Z + 1ms
                1716857100000, // 2024-05-28T00:45:00Z
              ],
            },
            {
              name: 'RotationsPerSecond',
              type: FieldType.number,
              config: {
                unit: 'RPS',
              },
              values: [1, 2, 3, 4],
            },
          ],
          length: 4,
        },
        {
          name: 'child',
          refId: 'B',
          fields: [
            {
              name: 'name',
              type: FieldType.string,
              config: {},
              values: ['child'],
            },
          ],
          length: 1,
        },
      ];

      it('does nothing when any query with client cache disabled', () => {
        const cacheMap = new Map();
        const cache = new RelativeRangeCache(cacheMap);

        cache.set(requestDisabledCache, {
          data: expectedDataFrames,
        });

        expect(cacheMap.size).toBe(0);
      });

      it('set request/response pair', () => {
        const cacheData = {
          [generateRequestCacheId(request)]: {
            queries: cachedQueryInfo,
            range,
          },
        };
        const expectedCacheMap = new Map(Object.entries(cacheData));

        const cacheMap = new Map();
        const cache = new RelativeRangeCache(cacheMap);

        cache.set(request, {
          data: expectedDataFrames,
        });

        expect(cacheMap).toEqual(expectedCacheMap);
      });
    });
  });
});
