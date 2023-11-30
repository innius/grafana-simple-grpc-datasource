import React, { useState } from 'react';
import { Dimension, VariableQuery, VariableQueryType } from '../types';
import { DataSource } from '../datasource';
import { SelectableValue } from '@grafana/data';
import DimensionSettings from './DimensionSettings';
import { AsyncSelect, InlineField, Input, Select, InlineFieldRow } from '@grafana/ui';

const formatDefinition = (query: VariableQuery): string => {
  switch (query.queryType) {
    case VariableQueryType.metric:
      return query.dimensions.map((x) => `${x.key}=${x.value}`).join(';');
    case VariableQueryType.dimensionValue:
      return `dimension=${query.dimensionKey}&filter=${query.dimensionValueFilter}`;
  }
};

const VariableQueryEditor = (props: {
  query: VariableQuery;
  onChange: (query: VariableQuery, definition: string) => void;
  datasource: DataSource;
}) => {
  const dims: Dimension[] = [];
  const { datasource, onChange, query } = props;
  const [state, updateState] = useState({ ...query, dimensions: query.dimensions || dims });

  const onChangeQueryType = (qt?: VariableQueryType) => {
    const newState = { ...state, queryType: qt || VariableQueryType.metric };
    updateState(newState);
    onChange(newState, formatDefinition(newState));
  };

  const onDimensionsChange = (dimensions: Dimension[]) => {
    const newState = { ...state, dimensions: dimensions };
    updateState(newState);
    onChange(newState, formatDefinition(newState));
  };

  const onDimensionKeyChange = (key?: string) => {
    const newState = { ...state, dimensionKey: key || '' };
    updateState(newState);
    onChange(newState, formatDefinition(newState));
  };

  const loadDimensionKeys = (query: string): Promise<Array<SelectableValue<string>>> => {
    return datasource.listDimensionKeys(query, [])
  };

  const options: Array<SelectableValue<VariableQueryType>> = [
    {
      value: VariableQueryType.metric,
      label: 'Metric',
      description: 'the query selects metrics',
    },
    {
      value: VariableQueryType.dimensionValue,
      label: 'Dimension Value',
      description: 'the query selects dimension values',
    },
  ];

  function onDimensionValueFilterChange(filter: string) {
    const newState = { ...state, dimensionValueFilter: filter };
    updateState(newState);
    onChange(newState, formatDefinition(newState));
  }

  return (
    <>
      <InlineField label={'Query Type'} labelWidth={24}>
        <Select onChange={(x) => onChangeQueryType(x.value)} options={options} value={state.queryType} width={32} />
      </InlineField>
      {state.queryType === VariableQueryType.metric && (
        <DimensionSettings initState={state.dimensions || []} onChange={onDimensionsChange} datasource={datasource} />
      )}
      {state.queryType === VariableQueryType.dimensionValue && (
        <>
          <InlineFieldRow>
            <InlineField labelWidth={24} label={'Dimension Key'}>
              <AsyncSelect
                defaultOptions={true}
                value={{ label: state.dimensionKey, value: state.dimensionKey }}
                cacheOptions={false}
                loadOptions={loadDimensionKeys}
                onChange={(e) => onDimensionKeyChange(e.value)}
                width={32}
              />
            </InlineField>
            <InlineField label="Filter" labelWidth={20} tooltip={'filter dimension values'}>
              <Input
                width={40}
                onChange={(x) => onDimensionValueFilterChange(x.currentTarget.value)}
                value={state.dimensionValueFilter}
              />
            </InlineField>
          </InlineFieldRow>
        </>
      )}
    </>
  );
};

export default VariableQueryEditor;
