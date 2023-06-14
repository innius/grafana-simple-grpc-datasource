import {DataSource} from '../datasource';
import {DataSourceInstanceSettings, ScopedVars, TimeRange, TypedVariableModel} from '@grafana/data';
import {MyDataSourceOptions, MyQuery, QueryType } from '../types';
import {setTemplateSrv} from '@grafana/runtime';

describe('Datasource', () => {
    const settings = {
        jsonData: {
            apikey_authentication_enabled: true,
        },
    } as DataSourceInstanceSettings<MyDataSourceOptions>;

    const ds = new DataSource(settings);

    // setup a test template server to mimic the real implementation
    setTemplateSrv({
        getVariables(): TypedVariableModel[] {
            return [];
        },
        replace(target?: string, scopedVars?: ScopedVars, format?: string | Function): string {
            if (target === '$sensor') {
                return JSON.stringify(['a', 'b', 'c']);
            }
            if (target === '$view') {
                return 'changes';
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
            metrics: [{metricId: 'foo'}, {metricId: '$sensor'}],
        };
        const res = ds.applyTemplateVariables(query, {});
        it('be expanded to a new list of metrics', () => {
            expect(res.metrics).toHaveLength(4);
            expect(res.metrics).toEqual([{metricId: 'foo'}, {metricId: 'a'}, {metricId: 'b'}, {metricId: 'c'}]);
        });
    });
    describe('a query with no template variables for query options should', () => {
        const query: MyQuery = {
            queryType: QueryType.GetMetricAggregate,
            refId: 'A',
            metrics: [{metricId: 'foo'}], 
            queryOptions: {
              'view' : {label: '$view', value: '$view'},
              'foo' : {label: 'the bar value', value: 'bar'},
            }
        };
        const res = ds.applyTemplateVariables(query, {});
        it('template variable should be replaced with the selected value', () => {
            expect(res.queryOptions).toEqual({
              'view' : {label: 'changes', value: 'changes'},
              'foo' : {label: 'the bar value', value: 'bar'},
            });
        });
    })

    describe('query display text should', () => {
        const input: MyQuery = {
            queryType: QueryType.GetMetricAggregate,
            refId: 'A',
            dimensions: [
                {key: 'dim1', value: '1', id: ''},
                {key: 'dim1', value: '2', id: ''},
            ],
            metrics: [{metricId: 'id1'}, {metricId: 'id2'}, {metricId: '$sensor'}],
        };
        const displayText = ds.getQueryDisplayText(input);
        it('be formatted to a nice string', () => {
            expect(displayText).toEqual('[dim1=1,dim1=2] id1&id2&$sensor');
        });
    });
});
