import React, { useState } from 'react';
import { Button, AsyncSelect, Icon } from '@grafana/ui';
import { Dimension } from '../types';
import { SelectableValue } from '@grafana/data';

interface DimensionRowProps {
  dimension: Dimension;
  onRemove: (dimension: Dimension) => void;
  onChange: (dimension: Dimension) => void;
  loadDimensions: (value: string) => Promise<Array<SelectableValue<string>>>;
  loadDimensionValues: (key: string, value: string) => Promise<Array<SelectableValue<string>>>;
}

const DimensionRow = (props: DimensionRowProps) => {
  const { dimension, loadDimensions, loadDimensionValues, onChange, onRemove } = props;
  const [state, updateState] = useState<Dimension>(dimension);

  const onChangeDimensionKey = (newKey?: string) => {
    onChangeDimension({ ...state, key: newKey || '' });
  };

  const onChangeDimensionValue = (newValue?: string) => {
    onChangeDimension({ ...state, value: newValue || '' });
  };

  const onChangeDimension = (dim: Dimension) => {
    updateState(dim);
    onChange(dim);
  };

  return (
    <div className="gf-form">
      <>
        <label className="gf-form-label width-10">Key</label>
        <AsyncSelect
          defaultOptions={true}
          value={{ label: dimension.key, value: dimension.key }}
          cacheOptions={false}
          loadOptions={loadDimensions}
          onChange={(e) => onChangeDimensionKey(e.value)}
        />
      </>
      <>
        <span className="gf-form-label width-10">Value</span>
        <AsyncSelect
          key={dimension.key}
          defaultOptions={true}
          value={{ label: dimension.value, value: dimension.value }}
          loadOptions={(query) => loadDimensionValues(dimension.key, query)}
          isSearchable={true}
          isClearable={true}
          allowCustomValue={true}
          onCreateOption={onChangeDimensionValue}
          onChange={(e) => onChangeDimensionValue(e?.value)}
        />
      </>
      <Button variant="secondary" onClick={(_) => onRemove(dimension)}>
        <Icon name="trash-alt" />
      </Button>
    </div>
  );
};

export default DimensionRow;
