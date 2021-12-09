import React, { PureComponent } from 'react';
import { css } from 'emotion';
import uniqueId from 'lodash/uniqueId';

import { Button, Icon, InlineFormLabel, LegacyForms, stylesFactory } from '@grafana/ui';
import { Dimension, Dimensions } from './types';
import { SelectableValue } from '@grafana/data';

import { Props } from './QueryEditor';

const { AsyncSelect } = LegacyForms;

export interface State {
  dimensions: Dimensions;
}

interface DimensionRowProps {
  dimension: Dimension;
  onReset: (id: string) => void;
  onRemove: (id: string) => void;
  onChange: (value: Dimension) => void;
  loadDimensions: (value: string) => Promise<Array<SelectableValue<string>>>;
  loadDimensionValues: (key: string, value: string) => Promise<Array<SelectableValue<string>>>;
  onBlur: () => void;
}

const getDimensionRowStyles = stylesFactory(() => {
  return {
    layout: css`
      display: flex;
      align-items: center;
      margin-bottom: 4px;
      > * {
        margin-left: 4px;
        margin-bottom: 0;
        height: 100%;
        &:first-child,
        &:last-child {
          margin-left: 0;
        }
      }
    `,
  };
});

const selectClass = css({
  minWidth: '160px',
});

const DimensionRow: React.FC<DimensionRowProps> = ({
  dimension,
  onBlur,
  onChange,
  onRemove,
  onReset,
  loadDimensions,
  loadDimensionValues,
}) => {
  const styles = getDimensionRowStyles();

  return (
    <div className={styles.layout}>
      <>
        <InlineFormLabel width={5}>Key</InlineFormLabel>
        <div className={selectClass}>
          <AsyncSelect
            width={12}
            defaultOptions={true}
            value={{ label: dimension.key, value: dimension.key }}
            cacheOptions={false}
            loadOptions={loadDimensions}
            onChange={e => onChange({ ...dimension, key: e.value || '' })}
          />
        </div>
      </>
      <>
        <InlineFormLabel width={5}>Value</InlineFormLabel>
        <AsyncSelect
          key={dimension.key}
          width={12}
          defaultOptions={true}
          value={{ label: dimension.value, value: dimension.value }}
          loadOptions={query => loadDimensionValues(dimension.key, query)}
          isSearchable={true}
          isClearable={true}
          onChange={e => onChange({ ...dimension, value: e ? e.value || '' : '' })}
        />
      </>
      <Button variant="secondary" size="xs" onClick={_e => onRemove(dimension.id)}>
        <Icon name="trash-alt" />
      </Button>
    </div>
  );
};

DimensionRow.displayName = 'DimensionRow';

export class DimensionSettings extends PureComponent<Props, State> {
  state: State = {
    dimensions: [],
  };

  constructor(props: Props) {
    super(props);
    const { dimensions } = this.props.query;
    this.state = {
      dimensions: dimensions || [],
    };
  }

  updateSettings = () => {
    const { dimensions } = this.state;

    this.props.onChange({
      ...this.props.query,
      dimensions: dimensions,
    });
  };

  onDimensionAdd = () => {
    this.setState(prevState => {
      // @ts-ignore
      return { dimensions: [...prevState.dimensions, { id: uniqueId(), key: '', value: '', configured: false }] };
    });
  };

  onDimensionChange = (dimensionIndex: number, value: Dimension) => {
    this.setState(({ dimensions }) => {
      return {
        dimensions: dimensions.map((item, index) => {
          if (dimensionIndex !== index) {
            return item;
          }
          return { ...value };
        }),
      };
    }, this.updateSettings);
  };

  onDimensionReset = (dimensionID: string) => {
    this.setState(({ dimensions }) => {
      return {
        dimensions: dimensions.map((h, i) => {
          if (h.id !== dimensionID) {
            return h;
          }
          return {
            ...h,
            value: '',
            configured: false,
          };
        }),
      };
    }, this.updateSettings);
  };

  onDimensionRemove = (dimensionId: string) => {
    this.setState(
      ({ dimensions }) => ({
        dimensions: dimensions.filter(h => h.id !== dimensionId),
      }),
      this.updateSettings
    );
  };

  getDimensionKeys = (query: string): Promise<Array<SelectableValue<string>>> => {
    return this.props.datasource.listDimensionKeys(query);
  };

  getDimensionValues = (key: string, query: string): Promise<Array<SelectableValue<string>>> => {
    return this.props.datasource.listDimensionsValues(key, query);
  };

  render() {
    const { dimensions } = this.state;
    return (
      <div className={'gf-form-group'}>
        <div className="gf-form">
          <h6>Dimensions</h6>
        </div>
        <div>
          {dimensions.map((dimension, i) => (
            <DimensionRow
              key={dimension.id}
              dimension={dimension}
              onChange={h => {
                this.onDimensionChange(i, h);
              }}
              onBlur={this.updateSettings}
              onRemove={this.onDimensionRemove}
              onReset={this.onDimensionReset}
              loadDimensions={this.getDimensionKeys}
              loadDimensionValues={this.getDimensionValues}
            />
          ))}
        </div>
        <div className="gf-form">
          <Button
            variant="secondary"
            icon="plus"
            onClick={e => {
              this.onDimensionAdd();
            }}
          >
            Add dimension
          </Button>
        </div>
      </div>
    );
  }
}

export default DimensionSettings;
