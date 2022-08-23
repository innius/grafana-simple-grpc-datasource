import { DataSource } from '../datasource';
import { DataSourceInstanceSettings, ScopedVars, TimeRange, VariableModel } from '@grafana/data';
import { Dimensions, MyDataSourceOptions, MyQuery, QueryType } from '../types';
import { setTemplateSrv } from '@grafana/runtime';

describe('Datasource', () => {
  const settings = {
    jsonData: {
      apikey_authentication_enabled: true,
    },
  } as DataSourceInstanceSettings<MyDataSourceOptions>;

  const ds = new DataSource(settings);

  // setup a test template server to mimic the real implementation
  setTemplateSrv({
    getVariables(): VariableModel[] {
      return [];
    },
    replace(target?: string, scopedVars?: ScopedVars, format?: string | Function): string {
      if (target === '$sensor') {
        return JSON.stringify(['a', 'b', 'c']);
      }
      return target || '';
    },
    containsTemplate: function (target?: string): boolean {
      throw new Error('Function not implemented.');
    },
    updateTimeRange: function (timeRange: TimeRange): void {
      throw new Error('Function not implemented.');
    },
  });

  describe('a query with a template variable should', () => {
    const query: MyQuery = {
      queryType: QueryType.GetMetricAggregate,
      refId: 'A',
      metrics: [{ metricId: 'foo' }, { metricId: '$sensor' }],
    };
    const res = ds.applyTemplateVariables(query, {});
    it('be expanded to a new list of metrics', () => {
      expect(res.metrics).toHaveLength(4);
      expect(res.metrics).toEqual([{ metricId: 'foo' }, { metricId: 'a' }, { metricId: 'b' }, { metricId: 'c' }]);
    });
  });

  describe('metricFind query should', () => {
    it('give no error if dimensions are not specified', () => {
      const dimensions = ds.parseDimensions('');
      expect(dimensions).toHaveLength(0);
    });
    it('parses a dimension string to dimensions', () => {
      const dimensions = ds.parseDimensions('machine=foo;sensor_type=discrete');
      expect(dimensions).toHaveLength(2);
      const expected: Dimensions = [
        { id: '', key: 'machine', value: 'foo' },
        { id: '', key: 'sensor_type', value: 'discrete' },
      ];
      expect(dimensions).toEqual(expected);
    });
  });

  describe('query display text should', () => {
    const input: MyQuery = {
      queryType: QueryType.GetMetricAggregate,
      refId: 'A',
      dimensions: [
        { key: 'dim1', value: '1', id: '' },
        { key: 'dim1', value: '2', id: '' },
      ],
      metrics: [{ metricId: 'id1' }, { metricId: 'id2' }, { metricId: '$sensor' }],
    };
    const displayText = ds.getQueryDisplayText(input);
    it('be formatted to a nice string', () => {
      expect(displayText).toEqual('[dim1=1,dim1=2] id1&id2&$sensor');
    });
  });
});
