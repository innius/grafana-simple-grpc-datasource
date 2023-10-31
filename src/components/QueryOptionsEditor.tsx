import React from 'react';

import { Checkbox, Select, InlineField } from '@grafana/ui';
import { SelectableValue } from '@grafana/data';

import { QueryOptionDefinition, QueryOptionDefinitions, QueryOptions, QueryOptionValue, OptionType } from '../types';

interface Props {
  // the configured options of a query
  options: QueryOptions;
  // the definition of the options
  optionDefinitions: QueryOptionDefinitions;
  onChange: (key: string, value?: QueryOptionValue) => void;
}

const QueryOptionsEditor = (props: Props) => {
  const { optionDefinitions } = props;
  return (
    <>
      {optionDefinitions.map((opt) => {
        const currentValue = props.options[opt.id] || {};
        return (
          <>
            <InlineField labelWidth={24} label={opt.label} tooltip={opt.description} required={opt.required}>
              {opt.type === OptionType.Enum ? (
                <EnumOptionField currentValue={currentValue} option={opt} onChange={(v) => props.onChange(opt.id, v)} />
              ) : opt.type === OptionType.Boolean ? (
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
  option: QueryOptionDefinition;
  currentValue?: QueryOptionValue;
  onChange: (opt?: QueryOptionValue) => void;
}

const EnumOptionField = (props: EnumOptionProps) => {
  const { option, currentValue, onChange } = props;
  let enumOptions: Array<SelectableValue<string>> = [];
  if (option.enumValues) {
    enumOptions = option.enumValues.map((value) => ({
      value: value.id,
      ...value,
    }));
  }
  const onCreateOption = (value: string) => {
    if (!value) {
      return;
    }

    onChange({ value: value, label: value });
  };

  const value = enumOptions.find((x) => x.value === currentValue?.value) || {};
  return (
    <>
      <Select
        value={value}
        options={enumOptions}
        onChange={(x) => onChange({ value: x.value, label: x.label })}
        onCreateOption={onCreateOption}
        isSearchable={true}
        allowCustomValue={true}
        invalid={option.required && !currentValue?.value}
        width={32}
      />
    </>
  );
};

export default QueryOptionsEditor;
