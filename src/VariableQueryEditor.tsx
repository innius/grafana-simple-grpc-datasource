import React, { useState } from 'react';
import { Dimension, migrateLegacyQuery, VariableQuery, VariableQueryType } from './types';
import { DataSource } from './datasource';
import { DataSourceVariableSupport } from '@grafana/data/types/variables';
import DimensionSettings from './components/DimensionSettings';
import { Select } from '@grafana/ui';
import { SelectableValue } from '@grafana/data';

export type Props = DataSourceVariableSupport<DataSource>;

const VariableQueryEditor = (props: {
  query: VariableQuery | string;
  onChange: (query: VariableQuery, definition: string) => void;
  datasource: DataSource;
}) => {
  const { datasource, onChange } = props;
  const query = migrateLegacyQuery(props.query);
  const [state, updateState] = useState(query);

  const formatDefinition = (query: VariableQuery): string => {
    return query.dimensions.map((x) => `${x.key}=${x.value}`).join(';');
  };

  const onDimensionsChange = (dimensions: Dimension[]) => {
    const newState = { ...state, dimensions: dimensions };
    updateState(newState);
    onChange(newState, formatDefinition(newState));
  };

  const options: Array<SelectableValue<VariableQueryType>> = [
    {
      value: VariableQueryType.metric,
      label: 'Metric',
      description: 'the query selects one or more metrics',
    },
  ];
  return (
    <>
      <div className="gf-form">
        <label className="gf-form-label width-10">Query Type</label>
        <Select onChange={() => {}} options={options} value={query.queryType} />
      </div>

      <DimensionSettings initState={state.dimensions} onChange={onDimensionsChange} datasource={datasource} />
    </>
  );
};

export default VariableQueryEditor;
