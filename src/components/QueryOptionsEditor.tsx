import React, { useState, useEffect } from 'react';
import { DataSource } from '../datasource';

import { Checkbox, Select, InlineField } from '@grafana/ui';
import { SelectableValue } from '@grafana/data';

import { QueryOption, QueryOptions, QueryType, OptionValue } from '../types';

interface Props {
  queryOptions: { [key: string]: OptionValue | undefined };
  queryType: QueryType;
  onChange: (key: string, value?: OptionValue) => void;
  datasource: DataSource;
}

const QueryOptionsEditor = (props: Props) => {
  const { queryType, datasource } = props;
  const [resourceData, setResourceData] = useState<QueryOptions>([]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const opts = await datasource.getQueryOptions(queryType);
        setResourceData(opts);
      } catch (error) {
        console.error('Error fetching resource data', error);
      }
    };
    fetchData();
  }, [datasource, queryType]);

  return (
    <>
      {resourceData.map((opt) => {
        const currentValue = props.queryOptions[opt.id] || {};
        return (
          <>
            <InlineField labelWidth={20} label={opt.label} tooltip={opt.description} required={opt.required}>
              {opt.type === 'Enum' ? (
                <EnumOptionField
                  currentValue={currentValue}
                  option={opt}
                  onChange={(v) => props.onChange(opt.id, v)}
                />
              ) : opt.type === 'Boolean' ? (
                <Checkbox
                  value={currentValue.value === 'true'}
                  onChange={(v) => {
                    const curr = v.currentTarget.checked.toString();
                    props.onChange(opt.id, { value: curr });
                  }}
                />
              ) : (
                <div />
              )}
            </InlineField>
          </>
        );
      })}
    </>
  );
};

interface EnumOptionProps {
  option: QueryOption;
  currentValue?: OptionValue;
  onChange: (opt?: OptionValue) => void;
}

const EnumOptionField = (props: EnumOptionProps) => {
  const { option, currentValue, onChange } = props;
  let options: Array<SelectableValue<string>> = [];
  if (option.enumValues) {
    options = option.enumValues.map((value) => ({
      label: value.label,
      description: value.description,
      value: value.id,
    }));
  }

  return (
    <>
      <Select
        value={options.find((x) => x.value === currentValue?.value) || null}
        options={options}
        onChange={(x) => onChange({value: x.value, label: x.label})}
        menuPlacement="bottom"
        isSearchable={true}
        width={16}
      />
    </>
  );
};

export default QueryOptionsEditor;
