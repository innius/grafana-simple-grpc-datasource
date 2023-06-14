import { CustomVariableSupport, DataQueryRequest, DataQueryResponse } from '@grafana/data';
import VariableQueryEditor from './components/VariableQueryEditor';

import { DataSource } from './datasource';
import { Observable, from } from 'rxjs';
import { map } from 'rxjs/operators';
import { VariableQuery } from 'types';

export class DatasourceVariableSupport extends CustomVariableSupport<DataSource> {
  editor = VariableQueryEditor;

  constructor(private datasource: DataSource) {
    super();
    this.query = this.query.bind(this);
  }

  async execute(query: VariableQuery) {
    return this.datasource.metricFindQuery(query);
  }

  query(request: DataQueryRequest<VariableQuery>): Observable<DataQueryResponse> {
    const result = this.execute(request.targets[0]);

    return from(result).pipe(map((data) => ({ data })));
  }
}
