import { DataFrame, DataQueryRequest, DataQueryResponse } from '@grafana/data';
import { Metadata, NextQuery, MyQuery } from 'types';

export function getNextQueries(request: DataQueryRequest<MyQuery>, rsp: DataQueryResponse) {
  if (rsp.data?.length) {
    const next: NextQuery[] = [];
    for (const frame of rsp.data as DataFrame[]) {
      const meta = frame.meta?.custom as Metadata;
      if (meta && meta.nextToken) {
        const query = request.targets.find((t) => t.refId === frame.refId);
        if (query) {
          next.push({
            ...query,
            nextToken: meta.nextToken,
          });
        }
      }
    }
    if (next.length) {
      return next;
    }
  }
  return undefined;
}
