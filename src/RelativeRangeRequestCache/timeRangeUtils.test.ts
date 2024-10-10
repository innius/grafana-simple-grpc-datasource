import { dateTime } from '@grafana/data';
import {
  getRefreshRequestRange,
  isCacheableTimeRange,
  isRelativeFromNow,
  isTimeRangeCoveringStart,
  minDateTime,
} from './timeRangeUtils';

describe('isCacheableTimeRange()', () => {
  it('returns true for TimeRange with relative time from before refresh minutes to now', () => {
    expect(
      isCacheableTimeRange({
        from: dateTime('2024-05-28T00:00:00Z'),
        to: dateTime('2024-05-28T00:30:00Z'),
        raw: {
          from: 'now-30m',
          to: 'now',
        },
      })
    ).toBe(true);
  });

  it('returns false for undefined TimeRange', () => {
    expect(isCacheableTimeRange(undefined)).toBe(false);
  });

  it('returns false for TimeRange with absolute time', () => {
    expect(
      isCacheableTimeRange({
        from: dateTime('2024-05-28T00:00:00Z'),
        to: dateTime('2024-05-28T00:15:00Z'),
        raw: {
          from: '2024-05-28T00:00:00Z',
          to: '2024-05-28T00:15:00Z',
        },
      })
    ).toBe(false);
  });

  it('returns false for TimeRange with relative time not greater than refresh minutes', () => {
    expect(
      isCacheableTimeRange({
        from: dateTime('2024-05-28T00:00:00Z'),
        to: dateTime('2024-05-28T00:15:00Z'),
        raw: {
          from: 'now-15m',
          to: 'now',
        },
      })
    ).toBe(false);
  });
});

describe('getRefreshRequestRange()', () => {
  it('returns time range with cache ending time as start', () => {
    const requestRange = {
      from: dateTime('2024-05-28T00:00:00Z'),
      to: dateTime('2024-05-28T02:00:00Z'),
      raw: {
        from: 'now-2h',
        to: 'now',
      },
    };

    const cacheRange = {
      from: dateTime('2024-05-28T00:00:00Z'),
      to: dateTime('2024-05-28T01:00:00Z'),
      raw: {
        from: 'now-2h',
        to: 'now',
      },
    };

    const resultTimeRange = getRefreshRequestRange(requestRange, cacheRange);

    expect(resultTimeRange.from.isSame(dateTime('2024-05-28T01:00:00Z'))).toBe(true);
    expect(resultTimeRange.to.isSame(dateTime('2024-05-28T02:00:00Z'))).toBe(true);
  });

  it('returns time range with refresh minutes when time ranges are the same', () => {
    const requestRange = {
      from: dateTime('2024-05-28T00:00:00Z'),
      to: dateTime('2024-05-28T02:00:00Z'),
      raw: {
        from: 'now-2h',
        to: 'now',
      },
    };

    const cacheRange = {
      from: dateTime('2024-05-28T00:00:00Z'),
      to: dateTime('2024-05-28T02:00:00Z'),
      raw: {
        from: 'now-2h',
        to: 'now',
      },
    };

    const resultTimeRange = getRefreshRequestRange(requestRange, cacheRange);

    expect(resultTimeRange.from.isSame(dateTime('2024-05-28T01:45:00Z'))).toBe(true);
    expect(resultTimeRange.to.isSame(dateTime('2024-05-28T02:00:00Z'))).toBe(true);
  });
});

describe('isTimeRangeCoveringStart()', () => {
  it('returns true when subject and object both starting at the same time.', () => {
    const range = {
      from: dateTime('2024-05-28T20:59:49.659Z'),
      to: dateTime('2024-05-28T21:29:49.659Z'),
      raw: { from: 'now-1h', to: 'now' },
    };
    const actual = isTimeRangeCoveringStart(range, range);

    expect(actual).toBe(true);
  });

  it('returns true when object is before subject and overlaps the subject start time', () => {
    const objectRange = {
      from: dateTime('2024-05-28T20:00:00Z'),
      to: dateTime('2024-05-28T22:00:00Z'),
      raw: { from: 'now-1h', to: 'now' },
    };
    const subjectRange = {
      from: dateTime('2024-05-28T21:00:00Z'),
      to: dateTime('2024-05-28T25:00:00Z'),
      raw: { from: 'now-1h', to: 'now' },
    };
    const actual = isTimeRangeCoveringStart(objectRange, subjectRange);

    expect(actual).toBe(true);
  });

  it('returns false when object is before subject but ends before the subject start time', () => {
    const objectRange = {
      from: dateTime('2024-05-28T20:00:00Z'),
      to: dateTime('2024-05-28T21:00:00Z'),
      raw: { from: 'now-1h', to: 'now' },
    };
    const subjectRange = {
      from: dateTime('2024-05-28T22:00:00Z'),
      to: dateTime('2024-05-28T25:00:00Z'),
      raw: { from: 'now-1h', to: 'now' },
    };
    const actual = isTimeRangeCoveringStart(objectRange, subjectRange);

    expect(actual).toBe(false);
  });

  it('returns false when object is before subject but ends at subject start time', () => {
    const objectRange = {
      from: dateTime('2024-05-28T20:00:00Z'),
      to: dateTime('2024-05-28T22:00:00Z'),
      raw: { from: 'now-1h', to: 'now' },
    };
    const subjectRange = {
      from: dateTime('2024-05-28T22:00:00Z'),
      to: dateTime('2024-05-28T25:00:00Z'),
      raw: { from: 'now-1h', to: 'now' },
    };
    const actual = isTimeRangeCoveringStart(objectRange, subjectRange);

    expect(actual).toBe(false);
  });
});

describe('minDateTime()', () => {
  it('returns the minimum DateTime', () => {
    expect(minDateTime(dateTime(0), dateTime(1), dateTime(2))).toEqual(dateTime(0));
  });
});

describe('isRelativeFromNow()', () => {
  it('returns true for TimeRange with relative times to now', () => {
    expect(
      isRelativeFromNow({
        from: 'now-1h',
        to: 'now',
      })
    ).toBe(true);
  });

  it('returns false for TimeRange with absolute times', () => {
    expect(
      isRelativeFromNow({
        from: '2024-05-28T00:00:00Z',
        to: '2024-05-28T01:00:00Z',
      })
    ).toBe(false);
  });

  it('returns false for TimeRange with absolute from time', () => {
    expect(
      isRelativeFromNow({
        from: '2024-05-28T00:00:00Z',
        to: 'now',
      })
    ).toBe(false);
  });

  it('returns false for TimeRange with absolute to time', () => {
    expect(
      isRelativeFromNow({
        from: 'now-1',
        to: '2024-05-28T00:00:00Z',
      })
    ).toBe(false);
  });
});
