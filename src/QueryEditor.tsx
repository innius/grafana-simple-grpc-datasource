import defaults from 'lodash/defaults';
import { lastValueFrom } from 'rxjs';
import React, { ChangeEvent, PureComponent } from 'react';
import { Select, AsyncMultiSelect, InlineField, Input } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';

import { DataSource } from './datasource';
import { defaultQuery, Dimension, MyDataSourceOptions, MyQuery, QueryType, OptionValue } from './types';
import { changeQueryType, QueryTypeInfo, queryTypeInfos } from 'queryInfo';
import DimensionSettings from './components/DimensionSettings';
import QueryOptionsEditor from './components/QueryOptionsEditor';
import { convertQuery } from './convert';

export type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  constructor(props: Props) {
    super(props);
  }

  onQueryTypeChange = (sel: SelectableValue<QueryType>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...changeQueryType(query, sel as QueryTypeInfo), queryOptions: undefined });
    onRunQuery();
  };

  onMetricChange(evt: Array<SelectableValue<string>>) {
    const { onChange, query, onRunQuery } = this.props;

    const m = evt.map((x) => ({ metricId: x.value }));
    onChange({ ...query, metrics: m });
    onRunQuery();
  }

  onAddMetric(metric?: string) {
    if (!metric) {
      return;
    }
    const { onChange, query, onRunQuery } = this.props;
    const { metrics } = query;
    onChange({ ...query, metrics: metrics?.concat({ metricId: metric }) || [{ metricId: metric }] });
    onRunQuery();
  }

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

  onQueryOptionsChange = (key: string, value?: OptionValue) => {
    const { onChange, query, onRunQuery } = this.props;
    const { queryOptions } = query;
    onChange({ ...query, queryOptions: { ...queryOptions, [key]: value || {} } });
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
    const key = this.props.query.dimensions?.map((x) => x.key + x.value).join();

    const selectedMetrics = query.metrics?.map((x) => ({ label: x.metricId, value: x.metricId }));
    // AsyncSelect is not perfect yet, see https://github.com/JedWatson/react-select/issues/1879 for an alternative solution
    return (
      <>
        <InlineField labelWidth={24} label="Query Type">
          <Select options={queryTypeInfos} value={currentQueryType} onChange={this.onQueryTypeChange} width={32} />
        </InlineField>
        <DimensionSettings
          initState={query.dimensions || []}
          datasource={this.props.datasource}
          onChange={this.onDimensionsChange}
        />
        <InlineField labelWidth={24} label="Metric">
          <AsyncMultiSelect
            width={32}
            key={key}
            defaultOptions={true}
            value={selectedMetrics}
            loadOptions={this.loadMetrics}
            onChange={(evt) => this.onMetricChange(evt)}
            onCreateOption={(x) => this.onAddMetric(x)}
            allowCustomValue={true}
            isSearchable={true}
          />
        </InlineField>
        <InlineField
          labelWidth={24}
          label="Display Name"
          tooltip={`use ${this.displayNameFields(query.dimensions)} for dynamic expressions`}
        >
          <Input value={query.displayName} type="text" width={32} onChange={this.onDisplayNameChange} />
        </InlineField>
            <QueryOptionsEditor
              onChange={this.onQueryOptionsChange}
              datasource={this.props.datasource}
              queryType={query.queryType}
              queryOptions={query.queryOptions || {}}
            />
        </>
    );
  }
}
