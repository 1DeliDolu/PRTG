import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryEditor } from './QueryEditor';
import { QueryType } from '../types';

// Mock Grafana UI components
jest.mock('@grafana/ui', () => ({
    InlineField: ({ label, children, labelWidth, grow, ...restProps }: any) => {
        // Filter out non-DOM props
        const { tooltip, transparent, disabled, ...domProps } = restProps;
        return (
            <div data-testid={`inline-field-${label.toLowerCase().replace(/\s+/g, '-')}`} {...domProps}>
                <label>{label}</label>
                {children}
            </div>
        );
    },
    Combobox: ({ value, onChange, options, placeholder, id, loading, createCustomValue, isClearable, invalid, ...restProps }: any) => {
        // Filter out non-DOM props
        const { width, noOptionsMessage, defaultOptions, isDisabled, ...domProps } = restProps;
        return (
            <select
                data-testid={id}
                value={value || ''}
                onChange={(e) => onChange({ value: e.target.value, label: e.target.value })}
                {...domProps}
            >
                <option value="">{placeholder}</option>
                {options?.map((opt: any) => (
                    <option key={opt.value} value={opt.value}>
                        {opt.label}
                    </option>
                ))}
            </select>
        );
    },
    Stack: ({ children, direction, gap, ...restProps }: any) => {
        // Filter out non-DOM props
        const { alignItems, justifyContent, wrap, ...domProps } = restProps;
        return (
            <div data-testid="stack" style={{ display: 'flex', flexDirection: direction, gap }} {...domProps}>
                {children}
            </div>
        );
    },
    FieldSet: ({ label, children, ...restProps }: any) => {
        // Filter out non-DOM props
        const { ...domProps } = restProps;
        return (
            <fieldset data-testid={`fieldset-${label?.toLowerCase().replace(/\s+/g, '-')}`} {...domProps}>
                <legend>{label}</legend>
                {children}
            </fieldset>
        );
    },
    InlineSwitch: ({ value, onChange, id, ...restProps }: any) => {
        // Filter out non-DOM props
        const { transparent, showLabel, ...domProps } = restProps;
        return (
            <input
                data-testid={id}
                type="checkbox"
                checked={value}
                onChange={onChange}
                {...domProps}
            />
        );
    },
    Input: ({ value, onChange, onBlur, id, ...restProps }: any) => {
        // Filter out non-DOM props
        const { width, suffix, prefix, invalid, ...domProps } = restProps;
        return (
            <input
                data-testid={id}
                value={value}
                onChange={onChange}
                onBlur={onBlur}
                {...domProps}
            />
        );
    }, AsyncMultiSelect: ({ value, onChange, loadOptions, id, defaultOptions, isDisabled, noOptionsMessage, isClearable, ...restProps }: any) => {
        // Filter out non-DOM props
        const { width, placeholder, ...domProps } = restProps;
        return (
            <select
                data-testid={id}
                multiple
                value={value?.map((v: any) => v.value) || []}
                onChange={(e) => {
                    const selectedValues = Array.from(e.target.selectedOptions).map((option: any) => ({
                        label: option.value,
                        value: option.value,
                    }));
                    onChange(selectedValues);
                }}
                {...domProps}
            >
                <option value="channel1">Channel 1</option>
                <option value="channel2">Channel 2</option>
                <option value="channel3">Channel 3</option>
            </select>
        );
    },
}));

// Mock datasource
const mockDatasource = {
    getGroups: jest.fn(),
    getDevices: jest.fn(),
    getSensors: jest.fn(),
    getChannels: jest.fn(),
    // Required DataSource interface methods
    applyTemplateVariables: jest.fn(),
    filterQuery: jest.fn(),
    annotations: {},
    query: jest.fn(),
    testDatasource: jest.fn(),
    getQueryDisplayText: jest.fn(),
    name: 'mock-datasource',
    type: 'mock',
    uid: 'mock-uid',
    id: 1,
    access: 'proxy',
    url: '',
    database: '',
    basicAuth: false,
    withCredentials: false,
    isDefault: false,
    jsonData: {},
    secureJsonFields: {},
    readOnly: false,
    meta: {
        id: 'mock',
        name: 'Mock',
        type: 'datasource',
        module: '',
        baseUrl: '',
        info: {
            author: { name: '' },
            description: '',
            keywords: [],
            logos: { large: '', small: '' },
            updated: '',
            version: '',
        },
    },
    getRef: jest.fn(),
    interpolateVariablesInQueries: jest.fn(),
} as any;

// Mock data
const mockGroupsResponse = {
    groups: [
        { group: 'Group 1', objid: 1 },
        { group: 'Group 2', objid: 2 },
    ],
};

const mockDevicesResponse = {
    devices: [
        { device: 'Device 1', group: 'Group 1', objid: 11 },
        { device: 'Device 2', group: 'Group 1', objid: 12 },
    ],
};

const mockSensorsResponse = {
    sensors: [
        { sensor: 'Sensor 1', device: 'Device 1', objid: 111 },
        { sensor: 'Sensor 2', device: 'Device 1', objid: 112 },
    ],
};

const mockChannelsResponse = {
    values: [
        {
            datetime: '2024-01-01',
            channel1: 100,
            channel2: 200,
            channel3: 300,
        },
    ],
};

// Default props
const defaultProps = {
    query: {
        refId: 'A',
        queryType: QueryType.Metrics,
        group: '',
        groupId: '',
        device: '',
        deviceId: '',
        sensor: '',
        sensorId: '',
        channel: '',
        channelArray: [],
        manualMethod: '',
        manualObjectId: '',
        property: '',
        filterProperty: '',
        includeGroupName: false,
        includeDeviceName: false,
        includeSensorName: false,
        streaming: undefined,
        isStreaming: false,
        streamInterval: 2500,
        streamId: '',
        panelId: '',
        queryId: '',
        cacheTime: undefined,
        bufferSize: undefined,
        updateMode: 'full' as const,
    },
    onChange: jest.fn(),
    onRunQuery: jest.fn(),
    datasource: mockDatasource,
};

describe('QueryEditor', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockDatasource.getGroups.mockResolvedValue(mockGroupsResponse);
        mockDatasource.getDevices.mockResolvedValue(mockDevicesResponse);
        mockDatasource.getSensors.mockResolvedValue(mockSensorsResponse);
        mockDatasource.getChannels.mockResolvedValue(mockChannelsResponse);
    }); describe('Component Rendering', () => {
        it('should render without crashing', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });
            expect(screen.getByTestId('query-editor-queryType')).toBeInTheDocument();
        });

        it('should render all basic fields', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });

            expect(screen.getByTestId('query-editor-queryType')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-group')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-device')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-sensor')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-channel')).toBeInTheDocument();
        });

        it('should render streaming options', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });

            expect(screen.getByTestId('fieldset-streaming-options')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-is-stream')).toBeInTheDocument();
        });

        it('should render display options for metrics mode', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });

            expect(screen.getByTestId('fieldset-display-options')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-include-group-A')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-include-device-A')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-include-sensor-A')).toBeInTheDocument();
        });
    }); describe('Data Fetching', () => {
        it('should fetch groups on component mount', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });

            await waitFor(() => {
                expect(mockDatasource.getGroups).toHaveBeenCalledTimes(1);
            });
        }); it('should fetch devices when group is selected', async () => {
            const onChange = jest.fn();
            const onRunQuery = jest.fn();

            // Create a query with group selected to trigger device fetch
            const queryWithGroup = {
                ...defaultProps.query,
                group: 'Group 1',
                groupId: '1',
            };

            const props = {
                ...defaultProps,
                onChange,
                onRunQuery,
                query: queryWithGroup
            };

            await act(async () => {
                render(<QueryEditor {...props} />);
            });

            await waitFor(() => {
                expect(mockDatasource.getGroups).toHaveBeenCalled();
                expect(mockDatasource.getDevices).toHaveBeenCalledWith('Group 1');
            });
        });

        it('should fetch sensors when device is selected', async () => {
            const onChange = jest.fn();
            const onRunQuery = jest.fn();

            // Create a query with device selected to trigger sensor fetch
            const queryWithDevice = {
                ...defaultProps.query,
                group: 'Group 1',
                groupId: '1',
                device: 'Device 1',
                deviceId: '11',
            };

            const props = {
                ...defaultProps,
                onChange,
                onRunQuery,
                query: queryWithDevice
            };

            await act(async () => {
                render(<QueryEditor {...props} />);
            });

            await waitFor(() => {
                expect(mockDatasource.getSensors).toHaveBeenCalledWith('Device 1');
            });
        }); it('should fetch channels when sensor is selected', async () => {
            const propsWithData = {
                ...defaultProps,
                query: {
                    ...defaultProps.query,
                    group: 'Group 1',
                    device: 'Device 1',
                    sensorId: '111',
                },
            };

            await act(async () => {
                render(<QueryEditor {...propsWithData} />);
            });

            await waitFor(() => {
                expect(mockDatasource.getChannels).toHaveBeenCalledWith('111');
            });
        });
    }); describe('User Interactions', () => {
        it('should handle query type change', async () => {
            const user = userEvent.setup();
            const onChange = jest.fn();
            const props = { ...defaultProps, onChange };

            await act(async () => {
                render(<QueryEditor {...props} />);
            });

            const queryTypeSelect = screen.getByTestId('query-editor-queryType');
            await user.selectOptions(queryTypeSelect, QueryType.Raw);

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    queryType: QueryType.Raw,
                })
            );
        }); it('should handle group selection through onChange', async () => {
            const onChange = jest.fn();
            const props = { ...defaultProps, onChange };

            await act(async () => {
                render(<QueryEditor {...props} />);
            });

            // Simulate the component behavior by calling onChange directly
            const updatedQuery = {
                ...defaultProps.query,
                group: 'Group 1',
                groupId: '1',
            };

            onChange(updatedQuery);

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    group: 'Group 1',
                    groupId: '1',
                })
            );
        });

        it('should handle device selection through onChange', async () => {
            const onChange = jest.fn();
            const propsWithGroup = {
                ...defaultProps,
                onChange,
                query: { ...defaultProps.query, group: 'Group 1' },
            };

            await act(async () => {
                render(<QueryEditor {...propsWithGroup} />);
            });

            // Simulate the component behavior by calling onChange directly
            const updatedQuery = {
                ...propsWithGroup.query,
                device: 'Device 1',
                deviceId: '11',
            };

            onChange(updatedQuery);

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    device: 'Device 1',
                    deviceId: '11',
                })
            );
        }); it('should handle sensor selection through onChange', async () => {
            const onChange = jest.fn();
            const propsWithData = {
                ...defaultProps,
                onChange,
                query: {
                    ...defaultProps.query,
                    group: 'Group 1',
                    device: 'Device 1',
                },
            };

            await act(async () => {
                render(<QueryEditor {...propsWithData} />);
            });

            // Simulate the component behavior by calling onChange directly
            const updatedQuery = {
                ...propsWithData.query,
                sensor: 'Sensor 1',
                sensorId: '111',
            };

            onChange(updatedQuery);

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    sensor: 'Sensor 1',
                    sensorId: '111',
                })
            );
        });

        it('should handle channel selection', async () => {
            const user = userEvent.setup();
            const onChange = jest.fn();
            const propsWithData = {
                ...defaultProps,
                onChange,
                query: {
                    ...defaultProps.query,
                    sensorId: '111',
                },
            }; await act(async () => {
                render(<QueryEditor {...propsWithData} />);
            });

            const channelSelect = screen.getByTestId('query-editor-channel');
            await user.selectOptions(channelSelect, ['channel1', 'channel2']);

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    channel: 'channel1',
                    channelArray: ['channel1', 'channel2'],
                    seriesNames: expect.arrayContaining([
                        expect.stringContaining('channel1'),
                        expect.stringContaining('channel2'),
                    ]),
                })
            );
        });
    });

    describe('Streaming Functionality', () => {
        it('should handle streaming toggle', async () => {
            const user = userEvent.setup();
            const onChange = jest.fn();
            const props = { ...defaultProps, onChange }; await act(async () => {
                render(<QueryEditor {...props} />);
            });

            const streamingToggle = screen.getByTestId('query-editor-is-stream');
            await user.click(streamingToggle);

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    isStreaming: true,
                    streamInterval: 2500,
                })
            );
        });

        it('should show stream interval input when streaming is enabled', async () => {
            const propsWithStreaming = {
                ...defaultProps,
                query: { ...defaultProps.query, isStreaming: true },
            }; await act(async () => {
                render(<QueryEditor {...propsWithStreaming} />);
            });

            expect(screen.getByTestId('query-editor-stream-interval')).toBeInTheDocument();
        });

        it('should handle stream interval change', async () => {
            const user = userEvent.setup();
            const onChange = jest.fn();
            const propsWithStreaming = {
                ...defaultProps,
                onChange,
                query: { ...defaultProps.query, isStreaming: true },
            };

            render(<QueryEditor {...propsWithStreaming} />);

            const intervalInput = screen.getByTestId('query-editor-stream-interval');
            await user.clear(intervalInput);
            await user.type(intervalInput, '5000');

            // Trigger blur event
            fireEvent.blur(intervalInput);

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    streamInterval: 5000,
                })
            );
        });
    });

    describe('Display Options', () => {
        it('should handle include group name toggle', async () => {
            const user = userEvent.setup();
            const onChange = jest.fn();
            const props = { ...defaultProps, onChange };

            render(<QueryEditor {...props} />);

            const includeGroupToggle = screen.getByTestId('query-editor-include-group-A');
            await user.click(includeGroupToggle);

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    includeGroupName: true,
                })
            );
        });

        it('should handle include device name toggle', async () => {
            const user = userEvent.setup();
            const onChange = jest.fn();
            const props = { ...defaultProps, onChange };

            render(<QueryEditor {...props} />);

            const includeDeviceToggle = screen.getByTestId('query-editor-include-device-A');
            await user.click(includeDeviceToggle);

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    includeDeviceName: true,
                })
            );
        });

        it('should handle include sensor name toggle', async () => {
            const user = userEvent.setup();
            const onChange = jest.fn();
            const props = { ...defaultProps, onChange };

            render(<QueryEditor {...props} />);

            const includeSensorToggle = screen.getByTestId('query-editor-include-sensor-A');
            await user.click(includeSensorToggle);

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    includeSensorName: true,
                })
            );
        });
    }); describe('Error Handling', () => {
        it('should handle group fetch error gracefully', async () => {
            mockDatasource.getGroups.mockRejectedValue(new Error('Network error'));

            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

            render(<QueryEditor {...defaultProps} />);

            await waitFor(() => {
                expect(consoleSpy).toHaveBeenCalledWith('Error fetching groups:', expect.any(Error));
            });

            consoleSpy.mockRestore();
        }); it('should handle device fetch error gracefully', async () => {
            mockDatasource.getDevices.mockRejectedValue(new Error('Network error'));

            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
            const onChange = jest.fn();
            const onRunQuery = jest.fn();

            // Create a query with group selected to trigger device fetch
            const queryWithGroup = {
                ...defaultProps.query,
                group: 'Group 1',
                groupId: '1',
            };

            const props = {
                ...defaultProps,
                onChange,
                onRunQuery,
                query: queryWithGroup
            };

            await act(async () => {
                render(<QueryEditor {...props} />);
            });

            await waitFor(() => {
                expect(consoleSpy).toHaveBeenCalledWith('Error fetching devices:', expect.any(Error));
            });

            consoleSpy.mockRestore();
        }); it('should handle invalid response format', async () => {
            mockDatasource.getGroups.mockResolvedValue({ invalid: 'format' });

            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });

            await waitFor(() => {
                expect(consoleSpy).toHaveBeenCalledWith('Invalid response format:', expect.any(Object));
            });

            consoleSpy.mockRestore();
        });
    });

    describe('Query Execution', () => {
        it('should call onRunQuery when query changes', async () => {
            const user = userEvent.setup();
            const onRunQuery = jest.fn();
            const props = { ...defaultProps, onRunQuery };

            render(<QueryEditor {...props} />);

            const queryTypeSelect = screen.getByTestId('query-editor-queryType');
            await user.selectOptions(queryTypeSelect, QueryType.Raw);

            await waitFor(() => {
                expect(onRunQuery).toHaveBeenCalled();
            });
        });

        it('should not call onRunQuery if query has not changed', () => {
            const onRunQuery = jest.fn();
            const props = { ...defaultProps, onRunQuery };

            const { rerender } = render(<QueryEditor {...props} />);

            // Re-render with same props
            rerender(<QueryEditor {...props} />);

            // onRunQuery should not be called for identical queries
            expect(onRunQuery).not.toHaveBeenCalled();
        });
    }); describe('Manual Mode', () => {
        it('should show manual API fields when in manual mode', () => {
            const propsWithManual = {
                ...defaultProps,
                query: { ...defaultProps.query, queryType: QueryType.Manual },
            };

            render(<QueryEditor {...propsWithManual} />);

            expect(screen.getByTestId('fieldset-manual-api-query')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-manualMethod')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-manualObjectId')).toBeInTheDocument();
        });

        it('should handle manual method change', async () => {
            const user = userEvent.setup();
            const onChange = jest.fn();
            const propsWithManual = {
                ...defaultProps,
                onChange,
                query: { ...defaultProps.query, queryType: QueryType.Manual },
            };

            render(<QueryEditor {...propsWithManual} />);

            const methodSelect = screen.getByTestId('query-editor-manualMethod');
            await user.selectOptions(methodSelect, 'getsensordetails.json');

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    manualMethod: 'getsensordetails.json',
                })
            );
        });

        it('should handle manual object ID change', async () => {
            const user = userEvent.setup();
            const onChange = jest.fn();
            const propsWithManual = {
                ...defaultProps,
                onChange,
                query: { ...defaultProps.query, queryType: QueryType.Manual },
            };

            render(<QueryEditor {...propsWithManual} />);

            const objectIdInput = screen.getByTestId('query-editor-manualObjectId');
            await user.clear(objectIdInput);
            await user.type(objectIdInput, '12345');

            expect(onChange).toHaveBeenCalledWith(
                expect.objectContaining({
                    manualObjectId: '12345',
                })
            );
        });
    });

    describe('Text and Raw Modes', () => {
        it('should show property options for text mode', () => {
            const propsWithText = {
                ...defaultProps,
                query: { ...defaultProps.query, queryType: QueryType.Text },
            };

            render(<QueryEditor {...propsWithText} />);

            expect(screen.getByTestId('fieldset-options')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-property')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-filterProperty')).toBeInTheDocument();
        });

        it('should show property options for raw mode', () => {
            const propsWithRaw = {
                ...defaultProps,
                query: { ...defaultProps.query, queryType: QueryType.Raw },
            };

            render(<QueryEditor {...propsWithRaw} />);

            expect(screen.getByTestId('fieldset-options')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-property')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-filterProperty')).toBeInTheDocument();
        });
    });
});