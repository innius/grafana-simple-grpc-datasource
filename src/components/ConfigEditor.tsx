import React from 'react';
import { InlineField, InlineLabel, Input, SecretInput, Slider, InlineSwitch } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { defaultDataSourceOptions, MyDataSourceOptions, MySecureJsonData } from 'types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> {
}

const ConfigEditor = ({ options, onOptionsChange }: Props) => {
    const jsonData = {
        ...defaultDataSourceOptions,
        ...options.jsonData,
    }
    const opts = {
        ...options,
        jsonData,
    }
    return (
        <div className="gf-form-group">
            <ServerSettings options={opts} onOptionsChange={onOptionsChange} />
            <StreamingSettings options={opts} onOptionsChange={onOptionsChange} />
            <SecureSettings options={opts} onOptionsChange={onOptionsChange} />
        </div>
    )
}

export default ConfigEditor;

const StreamingSettings = ({ options, onOptionsChange }: Props) => {
    const onStreamingToggle = (enableStreaming: boolean) => {
        const jsonData = {
            ...options.jsonData,
            enableStreaming,
        };
        onOptionsChange({ ...options, jsonData });
    };

    return (
        <div className="gf-form-group">
            <h6>Streaming Configuration</h6>
            <div className="gf-form">
                <InlineField
                    label="Enable Streaming Queries"
                    labelWidth={30}
                    tooltip="Enable streaming queries for this datasource. When enabled, query editors will show streaming configuration options."
                >
                    <InlineSwitch
                        value={options.jsonData.enableStreaming || false}
                        onChange={(event) => onStreamingToggle(event.currentTarget.checked)}
                    />
                </InlineField>
            </div>
        </div>
    );
};

const SecureSettings = ({ options, onOptionsChange }: Props) => {
    const onAPIKeyChange = (apikey: string) => {
        onOptionsChange({
            ...options,
            secureJsonData: {
                apiKey: apikey,
            },
        });
    };

    const onResetAPIKey = () => {
        onOptionsChange({
            ...options,
            secureJsonFields: {
                ...options.secureJsonFields,
                apiKey: false,
            },
            secureJsonData: {
                ...options.secureJsonData,
                apiKey: '',
            },
        })
    }
    return (
        <div className="gf-form-group">
            <h6>Authentication</h6>
            <div className="gf-form">
                <InlineLabel width={30}
                    tooltip="The API key for backend API authentication">
                    API Key
                </InlineLabel>
                <SecretInput
                    width={40}
                    value={options.secureJsonData?.apiKey}
                    isConfigured={options.secureJsonFields.apiKey}
                    placeholder={"enter your backend api key"}
                    onReset={onResetAPIKey}
                    onChange={(event) => onAPIKeyChange(event.currentTarget.value.trim())}
                />
            </div>
        </div>
    )
}
const ServerSettings = ({ options, onOptionsChange }: Props) => {
    const onEndpointChange = (endpoint: string) => {
        const jsonData = {
            ...options.jsonData,
            endpoint: endpoint,
        };
        onOptionsChange({ ...options, jsonData });
    };

    const updateMaxRetries = (maxRetries: number) => {
        const jsonData = {
            ...options.jsonData,
            max_retries: maxRetries,
        };
        onOptionsChange({ ...options, jsonData });
    };

    return (
        <div className="gf-form-group">
            <h6>Server Settings</h6>
            <div className="gf-form">
                <InlineField label="Endpoint" labelWidth={30}
                    tooltip={"Specify a complete HTTP URL (for example grpc.example.com:443)"}>
                    <Input width={40} placeholder="endpoint of the grpc server" value={options.jsonData.endpoint}
                        onChange={x => onEndpointChange(x.currentTarget.value)} />
                </InlineField>
            </div>
            <div className="gf-form">
                <InlineLabel width={30}
                    tooltip="The number of times a backend invocation is retried if rate limit is reached">
                    Max. Retries
                </InlineLabel>
                <div style={{ width: '300px' }}>
                    <Slider min={0} max={10} onChange={updateMaxRetries} value={options.jsonData.max_retries} />
                </div>
            </div>
        </div>

    )
}
