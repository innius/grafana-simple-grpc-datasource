import defaults from 'lodash/defaults';
import { lastValueFrom } from 'rxjs';
import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms, AsyncMultiSelect } from '@grafana/ui';
import { QueryEditorProps, Registry, SelectableValue } from '@grafana/data';

import { DataSource } from './datasource';
import { AggregateType, defaultQuery, Dimension, MyDataSourceOptions, MyQuery, QueryType } from './types';
import { changeQueryType, QueryTypeInfo, queryTypeInfos } from 'queryInfo';
import DimensionSettings from './components/DimensionSettings';
import { convertQuery } from './convert';

const { Select, FormField } = LegacyForms;

export type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export const aggReg = new Registry(() => [
  { id: AggregateType.AVERAGE, name: 'Average' },
  { id: AggregateType.COUNT, name: 'Count' },
  { id: AggregateType.MAXIMUM, name: 'Max' },
  { id: AggregateType.MINIMUM, name: 'Min' },
]);

export class QueryEditor extends PureComponent<Props> {
  constructor(props: Props) {
    super(props);
  }

  onQueryTypeChange = (sel: SelectableValue<QueryType>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange(changeQueryType(query, sel as QueryTypeInfo));
    onRunQuery();
  };

  onMetricChange(evt: Array<SelectableValue<string>>) {
    const { onChange, query, onRunQuery } = this.props;

    const m = evt.map((x) => ({ metricId: x.value }));
    onChange({ ...query, metrics: m });
    onRunQuery();
  }

  onAggregateChange = (item: SelectableValue<AggregateType>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, aggregateType: item && item.value });
    onRunQuery();
  };

  onDisplayNameChange = (item: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, displayName: item && item.target.value });
    onRunQuery();
  };

  onDimensionsChange = (dimensions: Dimension[]) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, dimensions: dimensions });
    onRunQuery();
  };

  loadMetrics = (value: string): Promise<Array<SelectableValue<string>>> => {
    const { datasource } = this.props;
    const { dimensions } = this.props.query;
    return lastValueFrom(datasource.listMetrics(dimensions || [], value));
  };
  // fields which can be used in display name expression
  displayNameFields = (dimensions?: Dimension[]) =>
    dimensions
      ?.map((x) => x.key)
      .concat(['metric', 'aggregate'])
      .map((x) => '{{' + x + '}}')
      .join();

  render() {
    const query = convertQuery(defaults(this.props.query, defaultQuery));
    const currentQueryType = queryTypeInfos.find((v) => v.value === query.queryType);
    const select = aggReg.selectOptions([query.aggregateType || '']);
    const key = this.props.query.dimensions?.map((x) => x.key + x.value).join();

    const selectedMetrics = query.metrics?.map((x) => ({ label: x.metricId, value: x.metricId }));
    // AsyncSelect is not perfect yet, see https://github.com/JedWatson/react-select/issues/1879 for an alternative solution
    return (
      <div className="gf-form-group">
        <>
          <div className="gf-form">
            <label className="gf-form-label width-10">Query Type</label>
            <Select
              options={queryTypeInfos}
              value={currentQueryType}
              onChange={this.onQueryTypeChange}
              placeholder="Select query type"
              menuPlacement="bottom"
            />
          </div>
          <DimensionSettings
            initState={query.dimensions || []}
            datasource={this.props.datasource}
            onChange={this.onDimensionsChange}
          />
          <>
            <div className={'gf-form'}>
              <label className="gf-form-label width-10">Metric</label>
              <AsyncMultiSelect
                key={key}
                defaultOptions={true}
                value={selectedMetrics}
                loadOptions={this.loadMetrics}
                onChange={(evt) => this.onMetricChange(evt)}
                isSearchable={true}
                isClearable={true}
              />
            </div>
            <div
              className={'gf-form'}
              hidden={currentQueryType ? currentQueryType.value !== QueryType.GetMetricAggregate : true}
            >
              <label className="gf-form-label width-10">Aggregate</label>
              <Select
                value={select.current} //TODO: improve this
                options={select.options as any}
                onChange={this.onAggregateChange}
              />
            </div>
          </>
          <>
            <FormField
              labelWidth={10}
              value={query.displayName}
              onChange={this.onDisplayNameChange}
              label="Display Name"
              type="text"
              tooltip={`use ${this.displayNameFields(query.dimensions)} for dynamic expressions`}
            />
          </>
        </>
      </div>
    );
  }
}
