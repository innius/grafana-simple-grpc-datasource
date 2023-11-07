import defaults from 'lodash/defaults';
import { lastValueFrom } from 'rxjs';
import React, { ChangeEvent, useState, useEffect, } from 'react';
import { Select, AsyncMultiSelect, InlineField, Input } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';

import { DataSource } from './datasource';
import { defaultQuery, Dimension, MyDataSourceOptions, MyQuery, QueryType, QueryOptionValue, QueryOptionDefinitions, OptionType, QueryOptions } from './types';
import { queryTypeInfos } from 'queryInfo';
import DimensionSettings from './components/DimensionSettings';
import QueryOptionsEditor from './components/QueryOptionsEditor';
import { convertQuery } from './convert';

export type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

const QueryEditor = (props: Props) => {
  const [query, setQuery] = useState(convertQuery(defaults(props.query, defaultQuery)));
  const { datasource } = props;

  const [queryType, setQueryType] = useState(query.queryType)
  const [queryOptionDefinitions, setQueryOptionDefinitions] = useState<QueryOptionDefinitions>([])
  const [queryOptions, setQueryOptions] = useState(query.queryOptions)

  const applyDefaultValues = (q: MyQuery, opts: QueryOptionDefinitions): QueryOptions => {
    const enums = opts.filter(opt => opt.type === OptionType.Enum);
    let defaultOptions = q.queryOptions || {}

    for (const opt of enums) {
      if (!defaultOptions[opt.id]) {
        const defaultValue = opt.enumValues.find(v => v.default);

        if (defaultValue) {
          defaultOptions[opt.id] = { value: defaultValue.id, label: defaultValue.label };
        }
      }
    }
    return defaultOptions
  }

  // set the default query option values for the current backend query options
  // load the query options from the backend for the current query type
  useEffect(() => {
    const fetchData = async () => {
      try {
        const opts = await datasource.getQueryOptionDefinitions(queryType, queryOptions!);
        setQueryOptionDefinitions(opts);
        setQuery(x => ({ ...x, queryOptions: applyDefaultValues(x, opts) }))
      } catch (error) {
        console.error('Error fetching resource data', error);
      }
    };
    fetchData();
  }, [datasource, queryType, queryOptions]);

  const updateAndRunQuery = (q: MyQuery) => {
    const { onChange, onRunQuery } = props;
    onChange(q);
    setQuery(q);
    onRunQuery();
  };

  const onQueryTypeChange = async (queryType: QueryType) => {
    setQueryType(queryType);
    updateAndRunQuery({ ...query, queryType: queryType });
  };

  const onMetricChange = (evt: Array<SelectableValue<string>>) => {
    const m = evt.map((x) => ({ metricId: x.value }));
    updateAndRunQuery({ ...query, metrics: m });
  };

  const onAddMetric = (metric?: string) => {
    if (!metric) {
      return;
    }
    const { metrics } = query;
    updateAndRunQuery({ ...query, metrics: metrics?.concat({ metricId: metric }) || [{ metricId: metric }] });
  };

  const onDisplayNameChange = (item: ChangeEvent<HTMLInputElement>) => {
    updateAndRunQuery({ ...query, displayName: item && item.target.value });
  };

  const onDimensionsChange = (dimensions: Dimension[]) => {
    updateAndRunQuery({ ...query, dimensions: dimensions });
  };

  const onQueryOptionsChange = (key: string, value?: QueryOptionValue) => {
    const updatedQueryOptions = { ...queryOptions, [key]: value || {} };

    setQueryOptions(updatedQueryOptions);

    updateAndRunQuery({ ...query, queryOptions: updatedQueryOptions });
  };

  const loadMetrics = (value: string): Promise<Array<SelectableValue<string>>> => {
    const { dimensions } = query;
    return lastValueFrom(datasource.listMetrics(dimensions || [], value));
  };

  // fields which can be used in display name expression
  const displayNameFields = (dimensions?: Dimension[]) =>
    dimensions
      ?.map((x) => x.key)
      .concat(['metric', 'aggregate'])
      .map((x) => '{{ ' + x + '}}')
      .join();

  const currentQueryType = queryTypeInfos.find((v) => v.value === query.queryType);
  const key = query.dimensions?.map((x) => x.key + x.value).join();

  const selectedMetrics = query.metrics?.map((x) => ({ label: x.metricId, value: x.metricId }));
  // AsyncSelect is not perfect yet, see https://github.com/JedWatson/react-select/issues/1879 for an alternative solution
  return (
    <>
      <InlineField labelWidth={24} label="Query Type">
        <Select options={queryTypeInfos} value={currentQueryType} onChange={x => onQueryTypeChange(x.value || QueryType.GetMetricAggregate)} width={32} />
      </InlineField>
      <DimensionSettings
        initState={query.dimensions || []}
        datasource={datasource}
        onChange={onDimensionsChange}
      />
      <InlineField labelWidth={24} label="Metric">
        <AsyncMultiSelect
          width={32}
          key={key}
          defaultOptions={true}
          value={selectedMetrics}
          loadOptions={loadMetrics}
          onChange={(evt) => onMetricChange(evt)}
          onCreateOption={(x) => onAddMetric(x)}
          allowCustomValue={true}
          isSearchable={true}
        />
      </InlineField>
      <InlineField
        labelWidth={24}
        label="Display Name"
        tooltip={`use ${displayNameFields(query.dimensions)} for dynamic expressions`}
      >
        <Input value={query.displayName} type="text" width={32} onChange={onDisplayNameChange} />
      </InlineField>
      <QueryOptionsEditor
        onChange={onQueryOptionsChange}
        options={query.queryOptions || {}}
        optionDefinitions={queryOptionDefinitions}
      />
    </>
  );
};

export default QueryEditor;
