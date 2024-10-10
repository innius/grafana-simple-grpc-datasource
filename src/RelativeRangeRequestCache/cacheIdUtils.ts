import { DataQueryRequest } from '@grafana/data';
import { MyQuery } from 'types';

export type RequestCacheId = string;

export function generateRequestCacheId(request: DataQueryRequest<MyQuery>): RequestCacheId {
  const {
    targets,
    range: {
      raw: { from },
    },
  } = request;

  return JSON.stringify([from, generateQueriesCacheId(targets)]);
}

type QueryCacheId = string;

export function generateQueriesCacheId(queries: MyQuery[]): QueryCacheId {
  const cacheIds = queries.map(generateQueryCacheId).sort();

  return JSON.stringify(cacheIds);
}

/**
 * Parse query to cache id.
 */
function generateQueryCacheId(query: MyQuery): QueryCacheId {
  const { datasource, queryType, metrics, dimensions, queryOptions, displayName } = query;

  /*
   * Stringify to preserve undefined optional properties
   * `Undefined` optional properties are preserved as `null`
   */
  return JSON.stringify([datasource, queryType, metrics, dimensions, queryOptions, displayName]);
}
