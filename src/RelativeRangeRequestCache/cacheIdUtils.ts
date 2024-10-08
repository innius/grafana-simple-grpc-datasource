import { DataQueryRequest } from '@grafana/data';
import { SitewiseQueriesUnion } from './types';

export type RequestCacheId = string;

export function generateSiteWiseRequestCacheId(request: DataQueryRequest<SitewiseQueriesUnion>): RequestCacheId {
  const {
    targets,
    range: {
      raw: { from },
    },
  } = request;

  return JSON.stringify([from, generateSiteWiseQueriesCacheId(targets)]);
}

type QueryCacheId = string;

export function generateSiteWiseQueriesCacheId(queries: SitewiseQueriesUnion[]): QueryCacheId {
  const cacheIds = queries.map(generateSiteWiseQueryCacheId).sort();

  return JSON.stringify(cacheIds);
}

/**
 * Parse query to cache id.
 */
function generateSiteWiseQueryCacheId(query: SitewiseQueriesUnion): QueryCacheId {
  const { datasource, queryType, metrics, dimensions, queryOptions, displayName } = query;

  /*
   * Stringify to preserve undefined optional properties
   * `Undefined` optional properties are preserved as `null`
   */
  return JSON.stringify([datasource, queryType, metrics, dimensions, queryOptions, displayName]);
}
