import defaults from 'lodash/defaults';
import { lastValueFrom } from 'rxjs';
import React, { ChangeEvent, useState, useEffect, } from 'react';
import { Select, AsyncMultiSelect, InlineField, Input } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';

import { DataSource } from './datasource';
import { defaultQuery, Dimension, MyDataSourceOptions, MyQuery, QueryType, QueryOptionValue, QueryOptionDefinitions, OptionType } from './types';
import { queryTypeInfos } from 'queryInfo';
import DimensionSettings from './components/DimensionSettings';
import QueryOptionsEditor from './components/QueryOptionsEditor';
import { convertQuery } from './convert';

export type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

// while loading the page: 
// 1. convert (a possible) legacy query to MyQuery type
// 2. async load the options for the current query type 
// 3. set the default values for the query based on the backend query options 
const QueryEditor = (props: Props) => {
  const [query, setQuery] = useState(convertQuery(defaults(props.query, defaultQuery)));
  const { datasource } = props;

  const [queryType, setQueryType] = useState(query.queryType)
  const [queryOptionDefinitions, setQueryOptionDefinitions] = useState<QueryOptionDefinitions>([])
  const [queryOptions, setQueryOptions] = useState(query.queryOptions)

  type queryOptionsType = { [key: string]: QueryOptionValue }

  // load the query options from the backend for the current query type
  useEffect(() => {
    const fetchData = async () => {
      try {
        const opts = await datasource.getQueryOptionDefinitions(queryType, queryOptions!);
        setQueryOptionDefinitions(opts);
      } catch (error) {
        console.error('Error fetching resource data', error);
      }
    };
    fetchData();
  }, [datasource, queryType, queryOptions]);

  // set the default query option values for the current backend query options
  useEffect(() => {
    const applyDefaultValues = (q: MyQuery, opts: QueryOptionDefinitions): queryOptionsType => {
      const enums = opts.filter(opt => opt.type === OptionType.Enum);

      let defaultOptions = q.queryOptions || {}
      enums.forEach(opt => {
        const defaultValue = opt.enumValues.find(v => v.default)
        if (!defaultOptions[opt.id] && defaultValue) {
          defaultOptions[opt.id] = { value: defaultValue.id, label: defaultValue.label }
        }
      })
      return defaultOptions
    }
    setQuery(x => ({ ...x, queryOptions: applyDefaultValues(x, queryOptionDefinitions) }))
  }, [queryOptionDefinitions])

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
    setQueryOptions({ ...queryOptions, [key]: value || {} });
    updateAndRunQuery({ ...query, queryOptions: { ...queryOptions, [key]: value || {} } });
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
