import { QueryType, MyQuery } from 'types';
import { generateQueriesCacheId as generateQueriesCacheId } from './cacheIdUtils';

function createQuery(id: number): MyQuery {
  return {
    metrics: [{ metricId: 'foo' }, { metricId: 'bar' }],
    dimensions: [{ id: 'baz', key: 'qux', value: 'quux' }],
    queryOptions: {},
    queryType: QueryType.GetMetricHistory,
    lastObservation: true,
    datasource: {
      type: 'innius-grpc-datasource',
      uid: 'mock-datasource-uid',
    },
    refId: `A-${id}`,
  };
}

describe('generateQueriesCacheId()', () => {
  it('parses Queries into cache Id', () => {
    const actualId = generateQueriesCacheId([createQuery(1), createQuery(2)]);
    const expectedId = JSON.stringify([
      '[{"type":"innius-grpc-datasource","uid":"mock-datasource-uid"},"GetMetricHistory",[{"metricId":"foo"},{"metricId":"bar"}],[{"id":"baz","key":"qux","value":"quux"}],{},null]',
      '[{"type":"innius-grpc-datasource","uid":"mock-datasource-uid"},"GetMetricHistory",[{"metricId":"foo"},{"metricId":"bar"}],[{"id":"baz","key":"qux","value":"quux"}],{},null]',
    ]);

    expect(actualId).toEqual(expectedId);
  });
});
