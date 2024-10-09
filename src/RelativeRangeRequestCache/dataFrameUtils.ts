import { FieldType, AbsoluteTimeRange, DataFrame } from '@grafana/data';
import { CachedQueryInfo, SitewiseQueriesUnion, isTimeOrderingQueryType, isTimeSeriesQueryType } from './types';
import { SiteWiseTimeOrder } from 'types';

function isRequestTimeDescending({ queryType, timeOrdering }: SitewiseQueriesUnion) {
  return isTimeOrderingQueryType(queryType) && timeOrdering === SiteWiseTimeOrder.DESCENDING;
}

/**
 * Trim cached query data frames based on the query type and time ordering for appending to the start of the data frame.
 *
 * @remarks
 * This function is used to trim the cached data frames based on the query type and time ordering
 * to ensure that the data frames are properly formatted for rendering.
 * For descending ordered data frames, it will return an empty data frame.
 * For property value queries, it will return an empty data frame.
 * For all other queries, it will return the trimmed data frame.
 *
 * @param cachedQueryInfos - Cached query infos to trim
 * @param cacheRange - Cache range to include
 * @returns Trimmed data frames
 */
export function trimCachedQueryDataFramesAtStart(
  cachedQueryInfos: CachedQueryInfo[],
  cacheRange: AbsoluteTimeRange
): DataFrame[] {
  return cachedQueryInfos.map((cachedQueryInfo) => {
    const { query, dataFrame } = cachedQueryInfo;
    const { queryType } = query;
    if (isRequestTimeDescending(query)) {
      // Descending ordering data frame are added at the end of the request to respect the ordering
      // See related function - trimCachedQueryDataFramesEnding()
      return {
        ...dataFrame,
        fields: [],
        length: 0,
      };
    }

    // Always refresh PropertyValue
    // if (queryType === QueryType.PropertyValue) {
    //   return {
    //     ...dataFrame,
    //     fields: [],
    //     length: 0,
    //   };
    // }

    if (isTimeSeriesQueryType(queryType)) {
      return trimTimeSeriesDataFrame({
        dataFrame: cachedQueryInfo.dataFrame,
        timeRange: cacheRange,
        lastObservation: cachedQueryInfo.query.lastObservation,
      });
    }

    // No trimming needed
    return dataFrame;
  });
}

/**
 * Trim cached query data frames based on the time ordering for appending to the end of the data frame.
 *
 * @remarks
 * This function is used to trim the cached data frames based on the time ordering
 * to ensure that the data frames are properly formatted for rendering.
 * For descending ordered data frames, it will return the trimmed data frame.
 * For all other queries, it will return an empty data frame.
 *
 * @param cachedQueryInfos - Cached query infos to trim
 * @param cacheRange - Cache range to include
 * @returns Trimmed data frames
 */
export function trimCachedQueryDataFramesEnding(
  cachedQueryInfos: CachedQueryInfo[],
  cacheRange: AbsoluteTimeRange
): DataFrame[] {
  return cachedQueryInfos
    .filter(({ query }) => isRequestTimeDescending(query))
    .map((cachedQueryInfo) => {
      return trimTimeSeriesDataFrameReversedTime({
        dataFrame: cachedQueryInfo.dataFrame,
        lastObservation: cachedQueryInfo.query.lastObservation,
        timeRange: cacheRange,
      });
    });
}

interface TrimParams {
  dataFrame: DataFrame;
  timeRange: AbsoluteTimeRange;
  lastObservation?: boolean;
}

/**
 * Trim the time series data frame to the specified time range.
 * @param trimParams - The parameters for trimming the data frame.
 * @param trimParams.dataFrame - The data frame to trim.
 * @param trimParams.timeRange - The time range to trim to.
 * @param trimParams.lastObservation - Whether to include the last observation in the range.
 * @returns The trimmed data frame.
 */
export function trimTimeSeriesDataFrame({
  dataFrame,
  timeRange: { from, to },
  lastObservation,
}: TrimParams): DataFrame {
  const { fields } = dataFrame;
  if (fields == null || fields.length === 0) {
    return {
      ...dataFrame,
      fields: [],
      length: 0,
    };
  }

  const timeField = fields.find((field) => field.name === 'time' && field.type === FieldType.time);
  if (timeField == null) {
    // return the original data frame if a time field cannot be found
    return dataFrame;
  }

  const timeValues = timeField.values;

  let fromIndex = timeValues.findIndex((time) => time > from); // from is exclusive
  if (fromIndex === -1) {
    // no time value within range; include no data in the slice
    fromIndex = timeValues.length;
  } else if (lastObservation) {
    // Keeps 1 extra data point before the range
    fromIndex = Math.max(fromIndex - 1, 0);
  }

  let toIndex = timeValues.findIndex((time) => time > to); // to is inclusive
  if (toIndex === -1) {
    // all time values before `to`
    toIndex = timeValues.length;
  }

  const trimmedFields = fields.map((field) => ({
    ...field,
    values: field.values.slice(fromIndex, toIndex),
  }));

  return {
    ...dataFrame,
    fields: trimmedFields,
    length: trimmedFields[0].values.length,
  };
}

/**
 * Trim the time series data frame to the specified time range where the time field is in reversed order.
 * @param trimParams - The parameters for trimming the data frame.
 * @param trimParams.dataFrame - The data frame to trim.
 * @param trimParams.timeRange - The time range to trim to.
 * @param trimParams.lastObservation - Whether to include the last observation in the range.
 * @returns The trimmed data frame.
 */
export function trimTimeSeriesDataFrameReversedTime({
  dataFrame,
  timeRange: { from, to },
  lastObservation,
}: TrimParams): DataFrame {
  const { fields } = dataFrame;
  if (fields == null || fields.length === 0) {
    return {
      ...dataFrame,
      fields: [],
      length: 0,
    };
  }

  const timeField = fields.find((field) => field.name === 'time' && field.type === FieldType.time);
  if (timeField == null) {
    // return the original data frame if a time field cannot be found
    return dataFrame;
  }

  // Copy before reverse in place
  const timeValues = [...timeField.values].reverse();

  let fromIndex = timeValues.findIndex((time) => time > from); // from is exclusive
  if (fromIndex === -1) {
    // no time value within range; include no data in the slice
    fromIndex = timeValues.length;
  } else if (lastObservation) {
    // Keeps 1 extra data point before the range
    fromIndex = Math.max(fromIndex - 1, 0);
  }

  let toIndex = timeValues.findIndex((time) => time > to); // to is inclusive
  if (toIndex === -1) {
    // all time values before `to`
    toIndex = timeValues.length;
  }

  const trimmedFields = fields.map((field) => {
    const dataValues = [...field.values].reverse().slice(fromIndex, toIndex);

    return {
      ...field,
      values: dataValues.reverse(),
    };
  });

  return {
    ...dataFrame,
    fields: trimmedFields,
    length: trimmedFields[0].values.length,
  };
}
