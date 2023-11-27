import React, { useState } from 'react';
import { Dimension } from '../types';
import uniqueId from 'lodash/uniqueId';
import DimensionRow from './DimensionRow';
import { DataSource } from '../datasource';
import { SelectableValue } from '@grafana/data';
import { Button } from '@grafana/ui';

interface Props {
  initState: Dimension[];
  onChange: (dims: Dimension[]) => void;
  datasource: DataSource;
}

const DimensionSettings = (props: Props) => {
  const { initState, onChange, datasource } = props;
  const [state, updateState] = useState<Dimension[]>(initState);

  const update = (newState: Dimension[]) => {
    updateState(newState);
    onChange(newState);
  };
  const addDimension = () => {
    const newDim = { id: uniqueId(), key: '', value: '' };
    update([...state, newDim]);
  };

  const changeDimension = (dim: Dimension) => {
    update([...state.filter((x) => x.id !== dim.id), dim]);
  };

  const removeDimension = (dim: Dimension) => {
    update(state.filter((x) => x.id !== dim.id));
  };

  const getDimensionKeys = (query: string): Promise<Array<SelectableValue<string>>> => {
    return datasource.listDimensionKeys(query, state);
  };

  const getDimensionValues = (key: string, query: string): Promise<Array<SelectableValue<string>>> => {
    return datasource.listDimensionsValues(key, query, state);
  };

  return (
    <>
      <div>
        {state.map((dimension, i) => (
          <DimensionRow
            key={dimension.id}
            dimension={dimension}
            onChange={changeDimension}
            onRemove={removeDimension}
            loadDimensions={getDimensionKeys}
            loadDimensionValues={getDimensionValues}
          />
        ))}
      </div>
      <div className="gf-form">
        <Button variant="secondary" icon="plus" onClick={addDimension}>
          Add dimension
        </Button>
      </div>
    </>
  );
};

export default DimensionSettings;
