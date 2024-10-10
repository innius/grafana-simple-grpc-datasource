import { DateTime, RawTimeRange, TimeRange, dateTime } from '@grafana/data';
import { DEFAULT_TIME_SERIES_REFRESH_MINUTES } from './constants';

/**
 * Check if the given TimeRange is cacheable. A TimeRange is cacheable if it is relative and has data 15 minutes ago.
 * @param TimeRange to check
 * @returns true if the TimeRange is cacheable, false otherwise
 */
export function isCacheableTimeRange(timeRange?: TimeRange): boolean {
  if (!timeRange) {
    return false;
  }

  const { from, to, raw } = timeRange;

  if (!isRelativeFromNow(raw)) {
    return false;
  }

  const defaultRefreshAgo = dateTime(to).subtract(DEFAULT_TIME_SERIES_REFRESH_MINUTES, 'minutes');
  if (!from.isBefore(defaultRefreshAgo)) {
    return false;
  }

  return true;
}

/**
 * Get the refresh TimeRange for a given TimeRange. The refresh TimeRange is the TimeRange that will be used to refresh the cache.
 *
 * @remarks
 * The refresh TimeRange is the TimeRange that will be used to refresh the cache. The TimeRange is usually 15 minutes ago until the end of the request.
 * Unless the cache data ends earlier than the 15 minutes refresh range, then the TimeRange starts from the end cache data until the end of the request.
 *
 * @param TimeRange to get the refresh TimeRange for
 * @param TimeRange cacheRange the TimeRange that will be used to refresh the cache
 * @returns TimeRange the refresh TimeRange
 */
export function getRefreshRequestRange(requestRange: TimeRange, cacheRange: TimeRange): TimeRange {
  const defaultRefreshAgo = dateTime(requestRange.to).subtract(DEFAULT_TIME_SERIES_REFRESH_MINUTES, 'minutes');
  const from = minDateTime(cacheRange.to, defaultRefreshAgo);

  return {
    from,
    to: requestRange.to,
    raw: requestRange.raw,
  };
}

/**
 * Checks if the subject TimeRange covers the start time of the object TimeRange.
 *
 * @param subjectRange - The TimeRange to check if it covers the start time of the object.
 * @param objectRange - The TimeRange to check if its start time is covered by the subject.
 * @returns True if the subject TimeRange covers the start time of the object, false otherwise.
 */
export function isTimeRangeCoveringStart(subjectRange: TimeRange, objectRange: TimeRange): boolean {
  const { from: subjectFrom, to: subjectTo } = subjectRange;
  const { from: objectFrom } = objectRange;

  /*
   * True if both time ranges start at the same time.
   *
   * Positive example (same from time):
   *   subject: <from>...
   *   object:  <from>...
   */
  if (objectFrom.isSame(subjectFrom)) {
    return true;
  }

  /*
   * True if subject starts before object starts and overlaps the object start time
   *
   * Positive example (subject from and to wrap around object from):
   *   subject: <from>......<to>
   *   object:  ......<from>....(disregard to)
   *
   * Negative example (subject from and to both before object):
   *   subject: <from>.<to>.......
   *   object:  ...........<from>.(disregard to)
   */
  if (subjectFrom.isBefore(objectFrom) && objectFrom.isBefore(subjectTo)) {
    return true;
  }

  return false;
}

export function minDateTime(firstDateTimes: DateTime, ...dateTimes: DateTime[]) {
  const minValue = Math.min(firstDateTimes.valueOf(), ...dateTimes.map((dateTime) => dateTime.valueOf()));
  return dateTime(minValue);
}

export function isRelativeFromNow(timeRange: RawTimeRange): boolean {
  const { from, to } = timeRange;

  if (typeof from !== 'string' || typeof to !== 'string') {
    return false;
  }

  return from.startsWith('now-') && to === 'now';
}
