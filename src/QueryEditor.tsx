import defaults from 'lodash/defaults';

import React, {ChangeEvent, PureComponent} from 'react';
import { InlineFormLabel, LegacyForms } from '@grafana/ui';
import { QueryEditorProps, Registry, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { AggregateType, defaultQuery, MyDataSourceOptions, MyQuery, QueryType } from './types';
import { changeQueryType, QueryTypeInfo, queryTypeInfos } from 'queryInfo';
import DimensionSettings from './QueryDimensions';

const { Select, AsyncSelect,FormField } = LegacyForms;

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

  onMetricChange = (sel: SelectableValue<string>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, metricName: sel.label, metricId: sel.value });
    onRunQuery();
  };

  onAggregateChange = (item: SelectableValue<AggregateType>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, aggregateType: item && item.value });
    onRunQuery();
  };

  onDisplayNameChange = (item : ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, displayName: item && item.target.value });
    onRunQuery();
  }

  loadMetrics = (value: string): Promise<Array<SelectableValue<string>>> => {
    const { datasource } = this.props;
    const { dimensions } = this.props.query;

    return datasource.listMetrics(dimensions || [], value);
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const currentQueryType = queryTypeInfos.find(v => v.value === query.queryType);
    const select = aggReg.selectOptions([query.aggregateType || '']);
    const key = this.props.query.dimensions?.map(x => x.key + x.value).join();
    // fields which can be used in display name expression
    const displayNameFields = query.dimensions?.map(x => x.key).concat(['metric', 'aggregate']).map(x => '{{' + x + '}}').join()

    // AsyncSelect is not perfect yet, see https://github.com/JedWatson/react-select/issues/1879 for an alternative solution
    return (
      <div className="gf-form-group">
        <>
          <div className="gf-form">
            <InlineFormLabel width={8}>Query Type</InlineFormLabel>
              <Select
                options={queryTypeInfos}
                value={currentQueryType}
                onChange={this.onQueryTypeChange}
                placeholder="Select query type"
                menuPlacement="bottom"
              />
          </div>
          <DimensionSettings {...this.props} />
          <>
            <div className={'gf-form'}>
              <InlineFormLabel width={8} tooltip={'start typing to query for metrics'}>
                Metric
              </InlineFormLabel>
              <AsyncSelect
                key={key}
                width={12}
                defaultOptions={true}
                value={{ label: query.metricName, value: query.metricId }}
                loadOptions={this.loadMetrics}
                noOptionsMessage={() => 'type to search for metrics'}
                onChange={this.onMetricChange}
                isSearchable={true}
                isClearable={true}
              />
            </div>
            <div
              className={'gf-form'}
              hidden={currentQueryType ? currentQueryType.value !== QueryType.GetMetricAggregate : true}
            >
              <InlineFormLabel width={8}>Aggregate</InlineFormLabel>
              <Select
                value={select.current} //TODO: improve this
                options={select.options as any}
                onChange={this.onAggregateChange}
              />
            </div>
          </>
          <>

            <FormField
                width={5}
                labelWidth={8}
                value={query.displayName}
                onChange={this.onDisplayNameChange}
                label="Display Name"
                type="text"
                tooltip={`use ${displayNameFields} for dynamic expressions`}
            />
          </>
        </>
      </div>
    );
  }
}
