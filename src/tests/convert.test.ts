import { MyQuery, QueryType } from 'types';
import { convertQuery } from '../convert';
describe('query-conversion', () => {
  describe('a query with deprecated aggregateType', () => {
    describe('average', () => {
      const legacyQuery: MyQuery = {
        refId: 'foo',
        queryType: QueryType.GetMetricAggregate,
        aggregateType: 'Average',
      };
      const query = convertQuery(legacyQuery);
      it('should convert to query with query options', () => {
        expect(query).toHaveProperty('queryOptions');
        expect(query.queryOptions).toHaveProperty('aggregateType');
        expect(query.queryOptions!.aggregateType).toEqual({ label: 'average', value: '0' });
      });
    });
    describe('min', () => {
      const legacyQuery: MyQuery = {
        refId: 'foo',
        queryType: QueryType.GetMetricAggregate,
        aggregateType: 'Min',
      };
      const query = convertQuery(legacyQuery);
      it('should convert to query with query options', () => {
        expect(query).toHaveProperty('queryOptions');
        expect(query.queryOptions).toHaveProperty('aggregateType');
        expect(query.queryOptions!.aggregateType).toEqual({ label: 'min', value: '2' });
      });
    });
    describe('max', () => {
      const legacyQuery: MyQuery = {
        refId: 'foo',
        queryType: QueryType.GetMetricAggregate,
        aggregateType: 'Max',
      };
      const query = convertQuery(legacyQuery);
      it('should convert to query with query options', () => {
        expect(query).toHaveProperty('queryOptions');
        expect(query.queryOptions).toHaveProperty('aggregateType');
        expect(query.queryOptions!.aggregateType).toEqual({ label: 'max', value: '1' });
      });
    });
    describe('count', () => {
      const legacyQuery: MyQuery = {
        refId: 'foo',
        queryType: QueryType.GetMetricAggregate,
        aggregateType: 'Count',
      };
      const query = convertQuery(legacyQuery);
      it('should convert to query with query options', () => {
        expect(query).toHaveProperty('queryOptions');
        expect(query.queryOptions).toHaveProperty('aggregateType');
        expect(query.queryOptions!.aggregateType).toEqual({ label: 'count', value: '3' });
      });
    });
  });
});
