import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { ConfigEditor } from './ConfigEditor';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

// Mock the timezone module
jest.mock('../timezone', () => ({
    timezoneOptions: [
        { label: 'UTC', value: 'UTC' },
        { label: 'America/New_York', value: 'America/New_York' },
    ],
}));

// Mock Grafana UI components
jest.mock('@grafana/ui', () => ({
    InlineField: ({ label, children }: any) => (
        <div data-testid={`field-${label.toLowerCase().replace(' ', '-')}`}>
            <label>{label}</label>
            {children}
        </div>
    ),
    Input: ({ id, onChange, value, ...props }: any) => (
        <input
            data-testid={id}
            onChange={onChange}
            value={value}
            {...props}
        />
    ),
    SecretInput: ({ id, onChange, value, onReset, isConfigured, ...props }: any) => (
        <div>
            <input
                data-testid={id}
                onChange={onChange}
                value={value}
                {...props}
            />
            <button data-testid={`${id}-reset`} onClick={onReset}>
                Reset
            </button>
        </div>
    ),
    Combobox: ({ options, value, onChange }: any) => (
        <select
            data-testid="timezone-combobox"
            value={value}
            onChange={(e) => onChange({ value: e.target.value })}
        >
            {options.map((option: any) => (
                <option key={option.value} value={option.value}>
                    {option.label}
                </option>
            ))}
        </select>
    ),
}));

describe('ConfigEditor', () => {
    const mockOnOptionsChange = jest.fn();

    const defaultProps: DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> = {
        onOptionsChange: mockOnOptionsChange,
        options: {
            id: 1,
            uid: 'test-uid',
            orgId: 1,
            name: 'Test DataSource',
            type: 'test-type',
            typeName: 'Test Type',
            typeLogoUrl: '',
            access: 'proxy',
            url: '',
            user: '',
            database: '',
            basicAuth: false,
            basicAuthUser: '',
            withCredentials: false,
            isDefault: false,
            jsonData: {
                path: '',
                cacheTime: 60,
                timezone: 'UTC',
            },
            secureJsonFields: {},
            secureJsonData: {},
            readOnly: false,
            version: 1,
        },
    };

    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('renders all form fields', () => {
        render(<ConfigEditor {...defaultProps} />);

        expect(screen.getByTestId('field-path')).toBeInTheDocument();
        expect(screen.getByTestId('field-api-key')).toBeInTheDocument();
        expect(screen.getByTestId('field-cache-time')).toBeInTheDocument();
        expect(screen.getByTestId('field-timezone')).toBeInTheDocument();
    });

    it('handles path change', () => {
        render(<ConfigEditor {...defaultProps} />);

        const pathInput = screen.getByTestId('config-editor-path');
        fireEvent.change(pathInput, { target: { value: 'test.server.com' } });

        expect(mockOnOptionsChange).toHaveBeenCalledWith({
            ...defaultProps.options,
            jsonData: {
                ...defaultProps.options.jsonData,
                path: 'test.server.com',
            },
        });
    });

    it('handles API key change', () => {
        render(<ConfigEditor {...defaultProps} />);

        const apiKeyInput = screen.getByTestId('config-editor-api-key');
        fireEvent.change(apiKeyInput, { target: { value: 'new-api-key' } });

        expect(mockOnOptionsChange).toHaveBeenCalledWith({
            ...defaultProps.options,
            secureJsonData: {
                ...defaultProps.options.secureJsonData,
                apiKey: 'new-api-key',
            },
        });
    });

    it('handles API key reset', () => {
        render(<ConfigEditor {...defaultProps} />);

        const resetButton = screen.getByTestId('config-editor-api-key-reset');
        fireEvent.click(resetButton);

        expect(mockOnOptionsChange).toHaveBeenCalledWith({
            ...defaultProps.options,
            secureJsonFields: {
                ...defaultProps.options.secureJsonFields,
                apiKey: false,
            },
            secureJsonData: {
                ...defaultProps.options.secureJsonData,
                apiKey: '',
            },
        });
    });

    it('handles cache time change with valid value', () => {
        render(<ConfigEditor {...defaultProps} />);

        const cacheTimeInput = screen.getByTestId('config-editor-cache-time');
        fireEvent.change(cacheTimeInput, { target: { value: '120' } });

        expect(mockOnOptionsChange).toHaveBeenCalledWith({
            ...defaultProps.options,
            jsonData: {
                ...defaultProps.options.jsonData,
                cacheTime: 120,
            },
        });
    });

    it('sets default cache time when empty value is provided', () => {
        render(<ConfigEditor {...defaultProps} />);

        const cacheTimeInput = screen.getByTestId('config-editor-cache-time');
        fireEvent.change(cacheTimeInput, { target: { value: '' } });

        expect(mockOnOptionsChange).toHaveBeenCalledWith({
            ...defaultProps.options,
            jsonData: {
                ...defaultProps.options.jsonData,
                cacheTime: 60,
            },
        });
    });

    it('does not update cache time for invalid values', () => {
        render(<ConfigEditor {...defaultProps} />);

        const cacheTimeInput = screen.getByTestId('config-editor-cache-time');
        fireEvent.change(cacheTimeInput, { target: { value: '5' } });

        expect(mockOnOptionsChange).not.toHaveBeenCalled();
    });

    it('handles timezone change', () => {
        render(<ConfigEditor {...defaultProps} />);

        const timezoneSelect = screen.getByTestId('timezone-combobox');
        fireEvent.change(timezoneSelect, { target: { value: 'America/New_York' } });

        expect(mockOnOptionsChange).toHaveBeenCalledWith({
            ...defaultProps.options,
            jsonData: {
                ...defaultProps.options.jsonData,
                timezone: 'America/New_York',
            },
        });
    });

    it('displays current values in form fields', () => {
        const propsWithData = {
            ...defaultProps,
            options: {
                ...defaultProps.options,
                jsonData: {
                    path: 'existing.server.com',
                    cacheTime: 180,
                    timezone: 'America/New_York',
                },
                secureJsonData: {
                    apiKey: 'existing-key',
                },
            },
        };

        render(<ConfigEditor {...propsWithData} />); expect(screen.getByTestId('config-editor-path')).toHaveValue('existing.server.com');
        expect(screen.getByTestId('config-editor-api-key')).toHaveValue('existing-key');
        expect(screen.getByTestId('config-editor-cache-time')).toHaveValue(180);
        expect(screen.getByTestId('timezone-combobox')).toHaveValue('America/New_York');
    });
});