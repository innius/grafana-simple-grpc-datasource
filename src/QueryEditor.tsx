import defaults from 'lodash/defaults';

import React, { PureComponent } from 'react';
import { InlineField, InlineFormLabel, LegacyForms } from '@grafana/ui';
import { QueryEditorProps, Registry, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { AggregateType, defaultQuery, MyDataSourceOptions, MyQuery, QueryType } from './types';
import { changeQueryType, QueryTypeInfo, queryTypeInfos } from 'queryInfo';
import DimensionSettings from './QueryDimensions';

const { Select, AsyncSelect } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export const aggReg = new Registry(() => [
  { id: AggregateType.AVERAGE, name: 'Average' },
  // { id: AggregateType.COUNT, name: 'Count', isValid: AnyTypeOK },
  { id: AggregateType.MAXIMUM, name: 'Max' },
  { id: AggregateType.MINIMUM, name: 'Min' },
  // { id: AggregateType.SUM, name: 'Sum', isValid: OnlyNumbers },
  // { id: AggregateType.STANDARD_DEVIATION, name: 'Stddev', description: 'Standard Deviation', isValid: OnlyNumbers },
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

  loadMetrics = (value: string): Promise<Array<SelectableValue<string>>> => {
    const { datasource } = this.props;
    const { dimensions } = this.props.query;
    return datasource.listMetrics(dimensions || [], value);
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const queryTooltip = '';
    const currentQueryType = queryTypeInfos.find(v => v.value === query.queryType);
    const select = aggReg.selectOptions([query.aggregateType || '']);
    // AsyncSelect is not perfect yet, see https://github.com/JedWatson/react-select/issues/1879 for an alternative solution
    return (
      <div className="gf-form-group">
        <>
          <div className="gf-form">
            <InlineField label="Query type" labelWidth={14} grow={true} tooltip={queryTooltip}>
              <Select
                options={queryTypeInfos}
                value={currentQueryType}
                onChange={this.onQueryTypeChange}
                placeholder="Select query type"
                menuPlacement="bottom"
              />
            </InlineField>
          </div>
          <DimensionSettings {...this.props} />
          <>
            <div className={'gf-form'}>
              <InlineFormLabel width={5} tooltip={'start typing to query for metrics'}>
                Metric
              </InlineFormLabel>
              <AsyncSelect
                width={12}
                defaultOptions={false}
                value={{ label: query.metricName, value: query.metricId }}
                loadOptions={this.loadMetrics}
                noOptionsMessage={() => 'type to search for metrics'}
                onChange={this.onMetricChange}
              />
            </div>

            <div
              className={'gf-form'}
              hidden={currentQueryType ? currentQueryType.value !== QueryType.GetMetricAggregate : true}
            >
              <InlineFormLabel width={5}>Aggregate</InlineFormLabel>
              <Select
                value={select.current} //TODO: improve this
                options={select.options as any}
                onChange={this.onAggregateChange}
              />
            </div>
          </>
        </>
      </div>
    );
  }
}
