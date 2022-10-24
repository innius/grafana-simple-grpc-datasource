import React from 'react';
import {InlineField, InlineLabel, Input, SecretInput, Slider} from '@grafana/ui';
import {DataSourcePluginOptionsEditorProps} from '@grafana/data';
import {defaultDataSourceOptions, MyDataSourceOptions, MySecureJsonData} from './types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> {
}

const ConfigEditor = ({options, onOptionsChange}: Props) => {
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
            <ServerSettings options={opts} onOptionsChange={onOptionsChange}/>
            <SecureSettings options={opts} onOptionsChange={onOptionsChange}/>
        </div>
    )
}

export default ConfigEditor;

const SecureSettings = ({options, onOptionsChange}: Props) => {
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
        <>
            <label>Authentication</label>
            <div className="gf-form">
                <InlineLabel width={20}
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
        </>
    )
}
const ServerSettings = ({options, onOptionsChange}: Props) => {
    const onEndpointChange = (endpoint: string) => {
        const jsonData = {
            ...options.jsonData,
            endpoint: endpoint,
        };
        onOptionsChange({...options, jsonData});
    };

    const updateMaxRetries = (maxRetries: number) => {
        const jsonData = {
            ...options.jsonData,
            max_retries: maxRetries,
        };
        onOptionsChange({...options, jsonData});
    };

    return (
        <div className="gf-form-group">
            <div className="gf-form">
                <InlineField label="Endpoint" labelWidth={20}
                             tooltip={"Specify a complete HTTP URL (for example grpc.example.com:443)"}>
                    <Input width={40} placeholder="endpoint of the grpc server" value={options.jsonData.endpoint}
                           onChange={x => onEndpointChange(x.currentTarget.value)}/>
                </InlineField>
            </div>
            <div className="gf-form">
                <InlineLabel width={20}
                             tooltip="The number of times a backend invocation is retried if rate limit is reached">
                    Max. Retries
                </InlineLabel>
                <div style={{width: '300px'}}>
                    <Slider min={0} max={10} onChange={updateMaxRetries} value={options.jsonData.max_retries}/>
                </div>
            </div>
        </div>

    )
}
