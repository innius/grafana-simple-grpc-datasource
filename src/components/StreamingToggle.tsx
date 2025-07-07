import React, { useState } from 'react';
import { InlineField, Switch, Input, Button, Tooltip } from '@grafana/ui';

interface StreamingToggleProps {
  value: boolean | string | undefined;
  onChange: (value: boolean | string) => void;
  disabled?: boolean;
}

const StreamingToggle: React.FC<StreamingToggleProps> = ({ value, onChange, disabled = false }) => {
  const [useVariableMode, setUseVariableMode] = useState(() => {
    // Start in variable mode if value is a string (regardless of content)
    return typeof value === 'string';
  });

  const isStreamingEnabled = () => {
    if (typeof value === 'boolean') return value;
    if (typeof value === 'string') {
      // Handle variable references - for display purposes, assume true if it's a variable
      if (value.includes('$')) return true;
      return value.toLowerCase() === 'true';
    }
    return false;
  };

  const handleSwitchChange = (event: React.FormEvent<HTMLInputElement>) => {
    if (disabled) return;
    const checked = (event.target as HTMLInputElement).checked;
    onChange(checked);
  };

  const handleTextChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (disabled) return;
    const textValue = event.target.value;
    onChange(textValue);
  };

  const toggleMode = () => {
    if (disabled) return;
    if (useVariableMode) {
      // Switch to boolean mode
      onChange(isStreamingEnabled());
    } else {
      // Switch to variable mode
      onChange(typeof value === 'string' ? value : '$streaming');
    }
    setUseVariableMode(!useVariableMode);
  };

  const getModeButtonText = () => useVariableMode ? 'Switch to Toggle' : 'Use Variable';
  const getModeButtonTooltip = () => 
    useVariableMode 
      ? 'Switch to simple on/off toggle' 
      : 'Use Grafana template variable (e.g., $streaming)';

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
      <InlineField 
        label="Streaming" 
        labelWidth={16} 
        tooltip={useVariableMode 
          ? "Enter a Grafana template variable (e.g., $streaming) or static value (true/false)" 
          : "Enable if the Grafana query should stream data"
        }
      >
        {useVariableMode ? (
          <Input
            value={typeof value === 'string' ? value : '$streaming'}
            onChange={handleTextChange}
            width={20}
            disabled={disabled}
            placeholder="$streaming"
          />
        ) : (
          <Switch 
            onChange={handleSwitchChange} 
            value={isStreamingEnabled()} 
            disabled={disabled}
          />
        )}
      </InlineField>
      <Tooltip content={getModeButtonTooltip()}>
        <Button
          variant="secondary"
          size="sm"
          onClick={toggleMode}
          disabled={disabled}
          icon={useVariableMode ? "toggle-on" : "edit"}
        >
          {getModeButtonText()}
        </Button>
      </Tooltip>
    </div>
  );
};

export default StreamingToggle;
