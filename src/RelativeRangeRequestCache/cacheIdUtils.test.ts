import { QueryType, SiteWiseTimeOrder } from 'types';
import { generateSiteWiseQueriesCacheId } from './cacheIdUtils';
import { SitewiseQueriesUnion } from './types';

function createSiteWiseQuery(id: number): SitewiseQueriesUnion {
  return {
    metrics: [{ metricId: 'foo' }, { metricId: 'bar' }],
    dimensions: [{ id: 'baz', key: 'qux', value: 'quux' }],
    queryOptions: {},
    queryType: QueryType.GetMetricHistory,
    lastObservation: true,
    datasource: {
      type: 'grafana-iot-sitewise-datasource',
      uid: 'mock-datasource-uid',
    },
    refId: `A-${id}`,
    timeOrdering: SiteWiseTimeOrder.ASCENDING,
  };
}

describe('generateSiteWiseQueriesCacheId()', () => {
  it('parses SiteWise Queries into cache Id', () => {
    const actualId = generateSiteWiseQueriesCacheId([createSiteWiseQuery(1), createSiteWiseQuery(2)]);
    const expectedId = JSON.stringify([
      '[{"type":"grafana-iot-sitewise-datasource","uid":"mock-datasource-uid"},"GetMetricHistory",[{"metricId":"foo"},{"metricId":"bar"}],[{"id":"baz","key":"qux","value":"quux"}],{},null]',
      '[{"type":"grafana-iot-sitewise-datasource","uid":"mock-datasource-uid"},"GetMetricHistory",[{"metricId":"foo"},{"metricId":"bar"}],[{"id":"baz","key":"qux","value":"quux"}],{},null]',
    ]);

    expect(actualId).toEqual(expectedId);
  });

  // it("parses SiteWise Query properties in a stable fashion (disregard of the order queries and queries' properties are added)", () => {
  //   // Reversed order of properties
  //   const query1: SitewiseQueriesUnion = {
  //     timeOrdering: SiteWiseTimeOrder.ASCENDING,
  //     queryType: QueryType.GetMetricHistory,
  //     refId: 'A-1',
  //     datasource: {
  //       uid: 'mock-datasource-uid',
  //       type: 'grafana-iot-sitewise-datasource',
  //     },
  //     // maxPageAggregations: 1000,
  //     // flattenL4e: true,
  //     lastObservation: true,
  //     // resolution: SiteWiseResolution.Auto,
  //     // quality: SiteWiseQuality.ANY,
  //     // propertyAlias: 'mock-property-alias-1',
  //     // propertyId: 'mock-property-id-1',
  //     // assetIds: ['mock-asset-id-1'],
  //     // assetId: 'mock-asset-id-1',
  //     // responseFormat: SiteWiseResponseFormat.Table,
  //     // region: 'us-west-2',
  //     // queryType: QueryType.PropertyValueHistory,
  //   };
  //   const query2 = {
  //     ...query1,
  //     queryType: QueryType.PropertyValue,
  //   };

  //   const order1 = generateSiteWiseQueriesCacheId([query2, query1]);
  //   const order2 = generateSiteWiseQueriesCacheId([query1, query2]);
  //
  //   expect(order1).toEqual(order2);
  // });

  //   it('parses SiteWise Query with only required properties provided', () => {
  //     // With only required properties
  //     const query: SitewiseQueriesUnion = {
  //       refId: 'A-1',
  //       queryType: QueryType.ListAssets,
  //     };
  //     const actualId = generateSiteWiseQueriesCacheId([query]);
  //     const expectedId = JSON.stringify([
  //       '["ListAssets",null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null,null]',
  //     ]);
  //
  //     expect(actualId).toEqual(expectedId);
  //   });
});

// describe('generateSiteWiseRequestCacheId()', () => {
//   it('parses SiteWise Queries into cache Id', () => {
//     const request = {
//       requestId: 'mock-request-id',
//       interval: '5s',
//       intervalMs: 5000,
//       range: {
//         from: dateTime('2024-05-28T00:00:00Z'),
//         to: dateTime('2024-05-28T01:00:00Z'),
//         raw: {
//           from: 'now-15m',
//           to: 'now',
//         },
//       },
//       scopedVars: {},
//       targets: [createSiteWiseQuery(1), createSiteWiseQuery(2)],
//       timezone: 'browser',
//       app: 'dashboard',
//       startTime: 1716858000000,
//     };
//     const expectedId = JSON.stringify([
//       'now-15m',
//       JSON.stringify([
//         '["PropertyValueHistory","us-west-2","table","mock-asset-id-1",["mock-asset-id-1"],"mock-property-id-1","mock-property-alias-1","ANY","AUTO",true,true,1000,"grafana-iot-sitewise-datasource","mock-datasource-uid","ASCENDING",true,"mock-hierarchy-1","mock-model-1","ALL",["AVERAGE"],"DISASSOCIATED","aws/mock/disassociated"]',
//         '["PropertyValueHistory","us-west-2","table","mock-asset-id-2",["mock-asset-id-2"],"mock-property-id-2","mock-property-alias-2","ANY","AUTO",true,true,1000,"grafana-iot-sitewise-datasource","mock-datasource-uid","ASCENDING",true,"mock-hierarchy-2","mock-model-2","ALL",["AVERAGE"],"DISASSOCIATED","aws/mock/disassociated"]',
//       ]),
//     ]);
//
//     expect(generateSiteWiseRequestCacheId(request)).toEqual(expectedId);
//   });
// });
