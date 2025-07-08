import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import ConfigEditor from '../ConfigEditor';
import { MyDataSourceOptions, MySecureJsonData } from '../../types';

const mockOnOptionsChange = jest.fn();

const defaultOptions = {
    id: 1,
    uid: 'test-uid',
    orgId: 1,
    name: 'Test Datasource',
    type: 'test-datasource',
    typeName: 'Test Datasource',
    typeLogoUrl: '',
    access: 'proxy' as const,
    url: '',
    user: '',
    database: '',
    basicAuth: false,
    basicAuthUser: '',
    withCredentials: false,
    isDefault: false,
    jsonData: {} as MyDataSourceOptions,
    secureJsonFields: {},
    secureJsonData: {} as MySecureJsonData,
    version: 1,
    readOnly: false,
};

describe('ConfigEditor', () => {
    beforeEach(() => {
        mockOnOptionsChange.mockClear();
    });

    it('renders streaming toggle', () => {
        render(
            <ConfigEditor
                options={defaultOptions}
                onOptionsChange={mockOnOptionsChange}
            />
        );

        expect(screen.getByText('Enable Streaming Queries')).toBeInTheDocument();
        expect(screen.getByRole('switch')).toBeInTheDocument();
    });

    it('toggles streaming setting', () => {
        render(
            <ConfigEditor
                options={defaultOptions}
                onOptionsChange={mockOnOptionsChange}
            />
        );

        const toggle = screen.getByRole('switch');
        fireEvent.click(toggle);

        expect(mockOnOptionsChange).toHaveBeenCalledWith({
            ...defaultOptions,
            jsonData: {
                max_retries: 5, // default value gets merged
                enableStreaming: true,
            },
        });
    });

    it('shows streaming toggle as checked when enabled', () => {
        const optionsWithStreaming = {
            ...defaultOptions,
            jsonData: {
                enableStreaming: true,
                apikey_authentication_enabled: false,
            },
        };

        render(
            <ConfigEditor
                options={optionsWithStreaming}
                onOptionsChange={mockOnOptionsChange}
            />
        );

        const toggle = screen.getByRole('switch');
        expect(toggle).toBeChecked();
    });
});
