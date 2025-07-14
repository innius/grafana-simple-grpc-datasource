import React, { ChangeEvent } from 'react';
import { InlineField, Input, InlineFieldRow } from '@grafana/ui';
import { StreamingConfig } from '../types';

interface StreamingConfigEditorProps {
    config: StreamingConfig;
    onChange: (config: StreamingConfig) => void;
    disabled?: boolean;
}

const StreamingConfigEditor: React.FC<StreamingConfigEditorProps> = ({ config, onChange, disabled = false }) => {
    const onMaxBufferSizeChange = (event: ChangeEvent<HTMLInputElement>) => {
        const value = parseInt(event.target.value, 10);
        if (!isNaN(value)) {
            onChange({ ...config, maxBufferSize: value });
        }
    };

    const onLookBackPeriodChange = (event: ChangeEvent<HTMLInputElement>) => {
        const value = event.target.value;
        onChange({ ...config, lookBackPeriod: value });
    };

    return (
        <div style={{ marginTop: '8px', padding: '8px', border: '1px solid #444', borderRadius: '4px', backgroundColor: '#1f1f23' }}>
            <div style={{ marginBottom: '8px', fontSize: '14px', fontWeight: 'bold', color: '#fff' }}>
                Streaming Configuration
            </div>
            <InlineFieldRow>
                <InlineField
                    label="Max Buffer Size"
                    labelWidth={20}
                    tooltip="Maximum number of datapoints to keep in the streaming buffer. Higher values use more memory but provide more historical context."
                    disabled={disabled}
                >
                    <Input
                        type="number"
                        value={config.maxBufferSize}
                        onChange={onMaxBufferSizeChange}
                        min={100}
                        max={100000}
                        step={100}
                        width={16}
                        disabled={disabled}
                        placeholder="3600"
                    />
                </InlineField>
                <InlineField
                    label="Look-back Period (seconds)"
                    labelWidth={24}
                    tooltip="Initial historical data period in seconds to load when starting the stream. Set to 0 to start with no historical data."
                    disabled={disabled}
                >
                    <Input
                        type="string"
                        value={config.lookBackPeriod}
                        onChange={onLookBackPeriodChange}
                        min={0}
                        max={86400}
                        step={60}
                        width={16}
                        disabled={disabled}
                        placeholder="1h"
                    />
                </InlineField>
            </InlineFieldRow>
            <div style={{ fontSize: '12px', color: '#999', marginTop: '4px' }}>
                Buffer size affects memory usage. Look-back period determines initial historical data loaded.
            </div>
        </div>
    );
};

export default StreamingConfigEditor;
