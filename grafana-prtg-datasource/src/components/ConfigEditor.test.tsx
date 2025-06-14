import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { ConfigEditor } from './ConfigEditor';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

// Mock Grafana UI components
jest.mock('@grafana/ui', () => ({
    InlineField: ({ children, label }: any) => (
        <div data-testid={`inline-field-${label.toLowerCase().replace(' ', '-')}`}>
            <label>{label}</label>
            {children}
        </div>
    ),
    Input: ({ id, onChange, value, placeholder }: any) => (
        <input
            data-testid={id}
            onChange={onChange}
            value={value}
            placeholder={placeholder}
        />
    ),
    SecretInput: ({ id, onChange, onReset, value, isConfigured }: any) => (
        <div>
            <input
                data-testid={id}
                onChange={onChange}
                value={value}
            />
            <button data-testid={`${id}-reset`} onClick={onReset}>
                Reset
            </button>
        </div>
    ),
}));

describe('ConfigEditor', () => {
    const mockOnOptionsChange = jest.fn();
    
    const defaultProps: DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> = {
        onOptionsChange: mockOnOptionsChange,
        options: {
            jsonData: {
                path: 'test.server.com',
                cacheTime: 6000,
            },
            secureJsonFields: {
                apiKey: false,
            },
            secureJsonData: {
                apiKey: 'test-api-key',
            },
        },
    } as any;

    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('renders all input fields', () => {
        render(<ConfigEditor {...defaultProps} />);
        
        expect(screen.getByTestId('config-editor-path')).toBeInTheDocument();
        expect(screen.getByTestId('config-editor-api-key')).toBeInTheDocument();
        expect(screen.getByTestId('config-editor-cache-time')).toBeInTheDocument();
    });

    it('displays current values in inputs', () => {
        render(<ConfigEditor {...defaultProps} />);
        
        expect(screen.getByTestId('config-editor-path')).toHaveValue('test.server.com');
        expect(screen.getByTestId('config-editor-api-key')).toHaveValue('test-api-key');
        expect(screen.getByTestId('config-editor-cache-time')).toHaveValue('6000');
    });

    it('calls onOptionsChange when path changes', () => {
        render(<ConfigEditor {...defaultProps} />);
        
        const pathInput = screen.getByTestId('config-editor-path');
        fireEvent.change(pathInput, { target: { value: 'new.server.com' } });
        
        expect(mockOnOptionsChange).toHaveBeenCalledWith({
            ...defaultProps.options,
            jsonData: {
                ...defaultProps.options.jsonData,
                path: 'new.server.com',
            },
        });
    });

    it('calls onOptionsChange when API key changes', () => {
        render(<ConfigEditor {...defaultProps} />);
        
        const apiKeyInput = screen.getByTestId('config-editor-api-key');
        fireEvent.change(apiKeyInput, { target: { value: 'new-api-key' } });
        
        expect(mockOnOptionsChange).toHaveBeenCalledWith({
            ...defaultProps.options,
            secureJsonData: {
                apiKey: 'new-api-key',
            },
        });
    });

    it('calls onOptionsChange when cache time changes', () => {
        render(<ConfigEditor {...defaultProps} />);
        
        const cacheTimeInput = screen.getByTestId('config-editor-cache-time');
        fireEvent.change(cacheTimeInput, { target: { value: '3600' } });
        
        expect(mockOnOptionsChange).toHaveBeenCalledWith({
            ...defaultProps.options,
            jsonData: {
                ...defaultProps.options.jsonData,
                cacheTime: 3600,
            },
        });
    });

    it('resets API key when reset button is clicked', () => {
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

    it('handles missing jsonData gracefully', () => {
        const propsWithoutJsonData = {
            ...defaultProps,
            options: {
                ...defaultProps.options,
                jsonData: {} as MyDataSourceOptions,
            },
        };
        
        render(<ConfigEditor {...propsWithoutJsonData} />);
        
        expect(screen.getByTestId('config-editor-cache-time')).toHaveValue('6000');
    });
});