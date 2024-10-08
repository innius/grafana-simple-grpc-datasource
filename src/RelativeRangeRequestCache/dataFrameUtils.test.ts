import { DataFrame, FieldType, dateTime } from '@grafana/data';
import { QueryType } from 'types';
import {
  trimTimeSeriesDataFrame,
  trimTimeSeriesDataFrameReversedTime,
  trimCachedQueryDataFramesAtStart,
} from './dataFrameUtils';

describe('trimCachedQueryDataFrames', () => {
  const absolutionRange = {
    from: dateTime('2024-05-28T00:00:00Z').valueOf(),
    to: dateTime('2024-05-28T00:15:00Z').valueOf(),
  };

  const dataFrame: DataFrame = {
    name: 'Demo Turbine Asset 1',
    refId: 'A',
    fields: [
      {
        name: 'time',
        type: FieldType.time,
        config: {},
        values: [
          1716854400000, // 2024-05-28T00:00:00Z
          1716854400001, // 2024-05-28T00:00:00Z + 1ms
          1716855300000, // 2024-05-28T00:15:00Z
          1716855300001, // 2024-05-28T00:15:00Z + 1ms
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
  };

  it.each([QueryType.GetMetricHistory])('trims time series data of time-series type - "%s"', (queryType: QueryType) => {
    const cachedQueryInfo = {
      query: {
        queryType,
        refId: 'A',
      },
      dataFrame,
    };
    const expectedDataFrame: DataFrame = {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: FieldType.time,
          config: {},
          values: [
            1716854400001, // +1ms
            1716855300000, // 2024-05-28T00:15:00Z
          ],
        },
        {
          name: 'RotationsPerSecond',
          type: FieldType.number,
          config: {
            unit: 'RPS',
          },
          values: [2, 3],
        },
      ],
      length: 2,
    };
    const dataFrames = trimCachedQueryDataFramesAtStart([cachedQueryInfo], absolutionRange);

    expect(dataFrames).toHaveLength(1);
    expect(dataFrames).toContainEqual(expectedDataFrame);
  });

  it('keeps all data when all time values within range', () => {
    const cachedQueryInfo = {
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
              1716854400001, // 2024-05-28T00:00:00Z+1ms
              1716855300000, // 2024-05-28T00:15:00Z
            ],
          },
          {
            name: 'RotationsPerSecond',
            type: FieldType.number,
            config: {
              unit: 'RPS',
            },
            values: [1, 2],
          },
        ],
        length: 2,
      },
    };
    const dataFrames = trimCachedQueryDataFramesAtStart([cachedQueryInfo], absolutionRange);

    expect(dataFrames).toHaveLength(1);
    expect(dataFrames).toContainEqual(cachedQueryInfo.dataFrame);
  });

  it('includes no time series data when all time values are before start time', () => {
    const cachedQueryInfo = {
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
              1716854399999,
              1716854400000, // 2024-05-28T00:00:00Z
            ],
          },
          {
            name: 'RotationsPerSecond',
            type: FieldType.number,
            config: {
              unit: 'RPS',
            },
            values: [1, 2],
          },
        ],
        length: 2,
      },
    };
    const expectedDataFrame: DataFrame = {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: FieldType.time,
          config: {},
          values: [],
        },
        {
          name: 'RotationsPerSecond',
          type: FieldType.number,
          config: {
            unit: 'RPS',
          },
          values: [],
        },
      ],
      length: 0,
    };
    const dataFrames = trimCachedQueryDataFramesAtStart([cachedQueryInfo], absolutionRange);

    expect(dataFrames).toHaveLength(1);
    expect(dataFrames).toContainEqual(expectedDataFrame);
  });

  it('includes no time series data when all time values are after end time', () => {
    const cachedQueryInfo = {
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
              1716855300001, // 2024-05-28T00:15:00Z +1ms
              1716855300002,
            ],
          },
          {
            name: 'RotationsPerSecond',
            type: FieldType.number,
            config: {
              unit: 'RPS',
            },
            values: [1, 2],
          },
        ],
        length: 2,
      },
    };
    const expectedDataFrame: DataFrame = {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: FieldType.time,
          config: {},
          values: [],
        },
        {
          name: 'RotationsPerSecond',
          type: FieldType.number,
          config: {
            unit: 'RPS',
          },
          values: [],
        },
      ],
      length: 0,
    };
    const dataFrames = trimCachedQueryDataFramesAtStart([cachedQueryInfo], absolutionRange);

    expect(dataFrames).toHaveLength(1);
    expect(dataFrames).toContainEqual(expectedDataFrame);
  });
});

describe('trimTimeSeriesDataFrame()', () => {
  const timeRange = {
    from: dateTime('2024-05-28T00:00:00Z').valueOf(),
    to: dateTime('2024-05-28T00:15:00Z').valueOf(),
  };

  const dataFrame: DataFrame = {
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
  };

  const dataFrameDiff: DataFrame = {
    name: 'Demo Turbine Asset 1',
    refId: 'A',
    fields: [
      {
        name: 'time',
        type: FieldType.time,
        config: {},
        values: [
          1716854400001, // 2024-05-28T00:15:00Z + 1ms
        ],
      },
      {
        name: 'RotationsPerSecond',
        type: FieldType.number,
        config: {
          unit: 'RPS',
        },
        values: [2],
      },
    ],
    length: 4,
  };

  it('trims time series data frame', () => {
    const trimParams = {
      dataFrame,
      timeRange,
    };
    const expectedDataFrame: DataFrame = {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: FieldType.time,
          config: {},
          values: [
            1716854400001, // +1ms
            1716855300000, // 2024-05-28T00:15:00Z
          ],
        },
        {
          name: 'RotationsPerSecond',
          type: FieldType.number,
          config: {
            unit: 'RPS',
          },
          values: [2, 3],
        },
      ],
      length: 2,
    };
    const dataFrameResult = trimTimeSeriesDataFrame(trimParams);

    expect(dataFrameResult).toEqual(expectedDataFrame);
  });

  it('trims time series data frame with last observations', () => {
    const trimParams = {
      dataFrame,
      lastObservation: true,
      timeRange,
    };
    const expectedDataFrame: DataFrame = {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: FieldType.time,
          config: {},
          values: [
            1716854400000, // 2024-05-28T00:00:00Z
            1716854400001, // +1ms
            1716855300000, // 2024-05-28T00:15:00Z
          ],
        },
        {
          name: 'RotationsPerSecond',
          type: FieldType.number,
          config: {
            unit: 'RPS',
          },
          values: [1, 2, 3],
        },
      ],
      length: 3,
    };
    const dataFrameResult = trimTimeSeriesDataFrame(trimParams);

    expect(dataFrameResult).toEqual(expectedDataFrame);
  });

  it('trims diff time series data frame', () => {
    const trimParams = {
      dataFrame: dataFrameDiff,
      timeRange,
    };
    const expectedDataFrame: DataFrame = {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: FieldType.time,
          config: {},
          values: [
            1716854400001, // +1ms
          ],
        },
        {
          name: 'RotationsPerSecond',
          type: FieldType.number,
          config: {
            unit: 'RPS',
          },
          values: [2],
        },
      ],
      length: 1,
    };
    const dataFrameResult = trimTimeSeriesDataFrame(trimParams);

    expect(dataFrameResult).toEqual(expectedDataFrame);
  });
});

describe('trimTimeSeriesDataFrameReversed()', () => {
  const timeRange = {
    from: dateTime('2024-05-28T00:00:00Z').valueOf(),
    to: dateTime('2024-05-28T00:15:00Z').valueOf(),
  };

  const dataFrameDescending: DataFrame = {
    name: 'Demo Turbine Asset 1',
    refId: 'A',
    fields: [
      {
        name: 'time',
        type: FieldType.time,
        config: {},
        values: [
          1716855300001, // 2024-05-28T00:15:00Z + 1ms
          1716855300000, // 2024-05-28T00:15:00Z
          1716854400001, // 2024-05-28T00:00:00Z + 1ms
          1716854400000, // 2024-05-28T00:00:00Z
        ],
      },
      {
        name: 'RotationsPerSecond',
        type: FieldType.number,
        config: {
          unit: 'RPS',
        },
        values: [4, 3, 2, 1],
      },
    ],
    length: 4,
  };

  const dataFrameDescendingDiff: DataFrame = {
    name: 'Demo Turbine Asset 1',
    refId: 'A',
    fields: [
      {
        name: 'time',
        type: FieldType.time,
        config: {},
        values: [
          1716854400001, // 2024-05-28T00:00:00Z + 1ms
        ],
      },
      {
        name: 'RotationsPerSecond',
        type: FieldType.number,
        config: {
          unit: 'RPS',
        },
        values: [2],
      },
    ],
    length: 4,
  };

  it('trims descending time series data frame', () => {
    const trimParams = {
      dataFrame: dataFrameDescending,
      timeRange,
    };
    const expectedDataFrame: DataFrame = {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: FieldType.time,
          config: {},
          values: [
            1716855300000, // 2024-05-28T00:15:00Z
            1716854400001, // 2024-05-28T00:00:00Z+1ms
          ],
        },
        {
          name: 'RotationsPerSecond',
          type: FieldType.number,
          config: {
            unit: 'RPS',
          },
          values: [3, 2],
        },
      ],
      length: 2,
    };
    const dataFrame = trimTimeSeriesDataFrameReversedTime(trimParams);

    expect(dataFrame).toEqual(expectedDataFrame);
  });

  it('trims descending time series data with last observations of time-series type - "%s"', () => {
    const trimParams = {
      dataFrame: dataFrameDescending,
      lastObservation: true,
      timeRange,
    };
    const expectedDataFrame: DataFrame = {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: FieldType.time,
          config: {},
          values: [
            1716855300000, // 2024-05-28T00:15:00Z
            1716854400001, // 2024-05-28T00:00:00Z+1ms
            1716854400000, // 2024-05-28T00:00:00Z
          ],
        },
        {
          name: 'RotationsPerSecond',
          type: FieldType.number,
          config: {
            unit: 'RPS',
          },
          values: [3, 2, 1],
        },
      ],
      length: 3,
    };
    const dataFrame = trimTimeSeriesDataFrameReversedTime(trimParams);

    expect(dataFrame).toEqual(expectedDataFrame);
  });

  it('trims diff descending time series data frame', () => {
    const trimParams = {
      dataFrame: dataFrameDescendingDiff,
      timeRange,
    };
    const expectedDataFrame: DataFrame = {
      name: 'Demo Turbine Asset 1',
      refId: 'A',
      fields: [
        {
          name: 'time',
          type: FieldType.time,
          config: {},
          values: [
            1716854400001, // 2024-05-28T00:00:00Z+1ms
          ],
        },
        {
          name: 'RotationsPerSecond',
          type: FieldType.number,
          config: {
            unit: 'RPS',
          },
          values: [2],
        },
      ],
      length: 1,
    };
    const dataFrame = trimTimeSeriesDataFrameReversedTime(trimParams);

    expect(dataFrame).toEqual(expectedDataFrame);
  });
});
