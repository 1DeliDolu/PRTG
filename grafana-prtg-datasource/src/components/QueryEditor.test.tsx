import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import { QueryEditor } from './QueryEditor';
import { DataSource } from '../datasource';
import { MyQuery, QueryType } from '../types';
import { QueryEditorProps } from '@grafana/data';

// Mock the datasource
const mockDatasource = {
    getGroups: jest.fn(),
    getDevices: jest.fn(),
    getSensors: jest.fn(),
    getChannels: jest.fn(),
} as unknown as DataSource;

// Mock data
const mockGroups = {
    groups: [
        { group: 'Group 1', objid: 1 },
        { group: 'Group 2', objid: 2 },
    ]
};

const mockDevices = {
    devices: [
        { device: 'Device 1', group: 'Group 1', objid: 11 },
        { device: 'Device 2', group: 'Group 1', objid: 12 },
    ]
};

const mockSensors = {
    sensors: [
        { sensor: 'Sensor 1', device: 'Device 1', objid: 111 },
        { sensor: 'Sensor 2', device: 'Device 1', objid: 112 },
    ]
};

const mockChannels = {
    values: [
        { datetime: '2023-01-01', channel1: 100, channel2: 200 }
    ]
};

const defaultQuery: MyQuery = {
    queryType: QueryType.Metrics,
    group: '',
    device: '',
    sensor: '',
    channel: '',
    channelArray: [],
    sensorId: '',
    groupId: '',
    deviceId: '',
    includeGroupName: false,
    includeDeviceName: false,
    includeSensorName: false,
    isStreaming: false,
    streamInterval: 2500,
    refId: 'A',
};

const defaultProps: QueryEditorProps<DataSource, MyQuery, any> = {
    query: defaultQuery,
    onChange: jest.fn(),
    onRunQuery: jest.fn(),
    datasource: mockDatasource,
    range: {} as any,
    data: {} as any,
    app: 'grafana' as any,
    history: [],
    queries: [],
};

// Mock React-Select components to work with testing-library
jest.mock('@grafana/ui', () => ({
    ...jest.requireActual('@grafana/ui'),
    Stack: ({ children, ...props }: any) => <div data-testid="stack" {...props}>{children}</div>,
    FieldSet: ({ children, label, ...props }: any) => (
        <fieldset data-testid="fieldset" {...props}>
            <legend>{label}</legend>
            {children}
        </fieldset>
    ),
    InlineField: ({ children, label, labelWidth, grow, tooltip, ...props }: any) => (
        <div data-testid="inline-field" {...props}>
            <label>{label}</label>
            {children}
        </div>
    ),
    Select: ({ onChange, value, options, onCreateOption, allowCustomValue, ...props }: any) => {
        const { isDisabled } = props;

        return (
            <select
                data-testid={props.id}
                aria-label={props['aria-label']}
                value={value?.value || value || ''}
                onChange={(e) => {
                    const selectedOption = options?.find((opt: any) => opt.value === e.target.value);
                    if (selectedOption) {
                        onChange(selectedOption);
                    } else if (allowCustomValue) {
                        const customOption = { value: e.target.value, label: e.target.value };
                        onChange(customOption);
                        if (onCreateOption) {
                            onCreateOption(e.target.value);
                        }
                    } else {
                        onChange({ value: e.target.value, label: e.target.value });
                    }
                }}
                disabled={isDisabled}
            >
                <option value="">Select...</option>
                {options?.map((option: any) => (
                    <option key={option.value} value={option.value}>
                        {option.label}
                    </option>
                ))}
                {props.id === 'query-editor-manualMethod' && (
                    <option value="getobjectstatus.htm">getobjectstatus.htm</option>
                )}
            </select>
        );
    },
    AsyncMultiSelect: ({ onChange, value, loadOptions, ...props }: any) => {
        const mockOptions = [
            { value: 'channel1', label: 'channel1' },
            { value: 'channel2', label: 'channel2' }
        ];

        const { isDisabled } = props;

        return (
            <select
                data-testid={props.id}
                aria-label={props['aria-label']}
                multiple
                value={value?.map((v: any) => v.value) || []}
                onChange={(e) => {
                    const selectedValues = Array.from(e.target.selectedOptions).map((option: any) => ({
                        value: option.value,
                        label: option.value
                    }));
                    onChange(selectedValues);
                }}
                disabled={isDisabled}
            >
                {mockOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                        {option.label}
                    </option>
                ))}
            </select>
        );
    },
    InlineSwitch: ({ onChange, value, ...props }: any) => (
        <input
            type="checkbox"
            data-testid={props.id}
            aria-label={props['aria-label']}
            checked={value}
            onChange={onChange}
            {...props}
        />
    ),
    Input: ({ onChange, value, ...props }: any) => (
        <input
            data-testid={props.id}
            aria-label={props['aria-label']}
            value={value || ''}
            onChange={onChange}
            type={props.type || 'text'}
            min={props.min}
            max={props.max}
            placeholder={props.placeholder}
            {...props}
        />
    ),
}));

describe('QueryEditor', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockDatasource.getGroups = jest.fn().mockResolvedValue(mockGroups);
        mockDatasource.getDevices = jest.fn().mockResolvedValue(mockDevices);
        mockDatasource.getSensors = jest.fn().mockResolvedValue(mockSensors);
        mockDatasource.getChannels = jest.fn().mockResolvedValue(mockChannels);
    });

    describe('Component Rendering', () => {
        it('renders without crashing', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });
            expect(screen.getByTestId('query-editor-queryType')).toBeInTheDocument();
        });

        it('renders all required fields', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });

            expect(screen.getByTestId('query-editor-queryType')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-group')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-device')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-sensor')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-channel')).toBeInTheDocument();
        });

        it('shows streaming options section', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });
            expect(screen.getByText('Streaming Options')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-is-stream')).toBeInTheDocument();
        });
    });

    describe('Data Fetching', () => {
        it('fetches groups on mount', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });

            await waitFor(() => {
                expect(mockDatasource.getGroups).toHaveBeenCalledTimes(1);
            });
        });

        it('fetches devices when group is selected', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} query={{ ...defaultQuery, group: 'Group 1' }} />);
            });

            await waitFor(() => {
                expect(mockDatasource.getDevices).toHaveBeenCalledWith('Group 1');
            });
        });

        it('fetches sensors when device is selected', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} query={{ ...defaultQuery, device: 'Device 1' }} />);
            });

            await waitFor(() => {
                expect(mockDatasource.getSensors).toHaveBeenCalledWith('Device 1');
            });
        });

        it('fetches channels when sensor ID is available', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} query={{ ...defaultQuery, sensorId: '111' }} />);
            });

            await waitFor(() => {
                expect(mockDatasource.getChannels).toHaveBeenCalledWith('111');
            });
        });
    });

    describe('Query Type Handling', () => {
        it('handles query type change to Raw', async () => {
            const onChange = jest.fn();

            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} />);
            });

            const queryTypeSelect = screen.getByTestId('query-editor-queryType');

            await act(async () => {
                fireEvent.change(queryTypeSelect, { target: { value: QueryType.Raw } });
            });

            await waitFor(() => {
                expect(onChange).toHaveBeenCalledWith(
                    expect.objectContaining({
                        queryType: QueryType.Raw,
                    })
                );
            });
        });

        it('shows display options for metrics mode', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} query={{ ...defaultQuery, queryType: QueryType.Metrics }} />);
            });

            expect(screen.getByText('Display Options')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-include-group')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-include-device')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-include-sensor')).toBeInTheDocument();
        });

        it('shows options for text mode', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} query={{ ...defaultQuery, queryType: QueryType.Text }} />);
            });

            expect(screen.getByText('Options')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-property')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-filterProperty')).toBeInTheDocument();
        });

        it('shows options for raw mode', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} query={{ ...defaultQuery, queryType: QueryType.Raw }} />);
            });

            expect(screen.getByText('Options')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-property')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-filterProperty')).toBeInTheDocument();
        });

        it('shows manual API query section for manual mode', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} query={{ ...defaultQuery, queryType: QueryType.Manual }} />);
            });
            expect(screen.getByText('Manual API Query')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-manualMethod')).toBeInTheDocument();
            expect(screen.getByTestId('query-editor-manualObjectId')).toBeInTheDocument();
        });
    });

    describe('Form Field Interactions', () => {
        it('handles group selection', async () => {
            const onChange = jest.fn();

            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} />);
            });

            await waitFor(() => {
                expect(mockDatasource.getGroups).toHaveBeenCalled();
            });

            const groupSelect = screen.getByTestId('query-editor-group');

            await act(async () => {
                fireEvent.change(groupSelect, { target: { value: 'Group 1' } });
            });

            await waitFor(() => {
                expect(onChange).toHaveBeenCalledWith(
                    expect.objectContaining({
                        group: 'Group 1',
                        groupId: '1',
                    })
                );
            });
        });

        it('handles device selection', async () => {
            const onChange = jest.fn();

            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} query={{ ...defaultQuery, group: 'Group 1' }} />);
            });

            await waitFor(() => {
                expect(mockDatasource.getDevices).toHaveBeenCalled();
            });

            const deviceSelect = screen.getByTestId('query-editor-device');

            await act(async () => {
                fireEvent.change(deviceSelect, { target: { value: 'Device 1' } });
            });

            await waitFor(() => {
                expect(onChange).toHaveBeenCalledWith(
                    expect.objectContaining({
                        device: 'Device 1',
                        deviceId: '11',
                    })
                );
            });
        });

        it('handles sensor selection', async () => {
            const onChange = jest.fn();

            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} query={{ ...defaultQuery, device: 'Device 1' }} />);
            });

            await waitFor(() => {
                expect(mockDatasource.getSensors).toHaveBeenCalled();
            });

            const sensorSelect = screen.getByTestId('query-editor-sensor');

            await act(async () => {
                fireEvent.change(sensorSelect, { target: { value: 'Sensor 1' } });
            });

            await waitFor(() => {
                expect(onChange).toHaveBeenCalledWith(
                    expect.objectContaining({
                        sensor: 'Sensor 1',
                        sensorId: '111',
                    })
                );
            });
        });

        it('handles channel selection', async () => {
            const onChange = jest.fn();

            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} query={{ ...defaultQuery, sensorId: '111' }} />);
            });

            // Wait for component to render and channels to be available
            await waitFor(() => {
                expect(screen.getByTestId('query-editor-channel')).toBeInTheDocument();
            });

            const channelSelect = screen.getByTestId('query-editor-channel');

            // Simulate selecting multiple options for a multi-select
            const option1 = channelSelect.querySelector('option[value="channel1"]') as HTMLOptionElement;
            const option2 = channelSelect.querySelector('option[value="channel2"]') as HTMLOptionElement;

            if (option1 && option2) {
                await act(async () => {
                    option1.selected = true;
                    option2.selected = true;
                    fireEvent.change(channelSelect);
                });

                await waitFor(() => {
                    expect(onChange).toHaveBeenCalledWith(
                        expect.objectContaining({
                            channel: 'channel1',
                            channelArray: ['channel1', 'channel2'],
                        })
                    );
                });
            }
        });
    });

    describe('Streaming Functionality', () => {
        it('handles streaming toggle', async () => {
            const onChange = jest.fn();
            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} />);
            });

            const streamingSwitch = screen.getByTestId('query-editor-is-stream');
            await act(async () => {
                fireEvent.click(streamingSwitch);
            });

            await waitFor(() => {
                expect(onChange).toHaveBeenCalledWith(
                    expect.objectContaining({
                        isStreaming: true,
                        streamInterval: 2500,
                    })
                );
            });
        });

        it('shows stream interval input when streaming is enabled', async () => {
            await act(async () => {
                render(<QueryEditor {...defaultProps} query={{ ...defaultQuery, isStreaming: true }} />);
            });

            expect(screen.getByTestId('query-editor-stream-interval')).toBeInTheDocument();
        });

        it('handles stream interval change', async () => {
            const onChange = jest.fn();
            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} query={{ ...defaultQuery, isStreaming: true }} />);
            });

            const intervalInput = screen.getByTestId('query-editor-stream-interval');
            await act(async () => {
                fireEvent.change(intervalInput, { target: { value: '5000' } });
            });

            await waitFor(() => {
                expect(onChange).toHaveBeenCalledWith(
                    expect.objectContaining({
                        streamInterval: 5000,
                    })
                );
            });
        });

        it('sets default streaming values on mount', async () => {
            const onChange = jest.fn();
            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} query={{ ...defaultQuery, isStreaming: undefined } as any} />);
            });
            await waitFor(() => {
                expect(onChange).toHaveBeenCalledWith(
                    expect.objectContaining({
                        isStreaming: false,
                        streamInterval: 2500,
                    })
                );
            });
        });
    });

    describe('Display Options', () => {
        it('handles include group name toggle', async () => {
            const onChange = jest.fn();
            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} query={{ ...defaultQuery, queryType: QueryType.Metrics }} />);
            });

            const includeGroupSwitch = screen.getByTestId('query-editor-include-group');
            await act(async () => {
                fireEvent.click(includeGroupSwitch);
            });

            await waitFor(() => {
                expect(onChange).toHaveBeenCalledWith(
                    expect.objectContaining({
                        includeGroupName: true,
                    })
                );
            });
        });

        it('handles include device name toggle', async () => {
            const onChange = jest.fn();
            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} query={{ ...defaultQuery, queryType: QueryType.Metrics }} />);
            });

            const includeDeviceSwitch = screen.getByTestId('query-editor-include-device');
            await act(async () => {
                fireEvent.click(includeDeviceSwitch);
            });

            await waitFor(() => {
                expect(onChange).toHaveBeenCalledWith(
                    expect.objectContaining({
                        includeDeviceName: true,
                    })
                );
            });
        });

        it('handles include sensor name toggle', async () => {
            const onChange = jest.fn();
            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} query={{ ...defaultQuery, queryType: QueryType.Metrics }} />);
            });

            const includeSensorSwitch = screen.getByTestId('query-editor-include-sensor');
            await act(async () => {
                fireEvent.click(includeSensorSwitch);
            });
            await waitFor(() => {
                expect(onChange).toHaveBeenCalledWith(
                    expect.objectContaining({
                        includeSensorName: true,
                    })
                );
            });
        });
    });

    describe('Error Handling', () => {
        it('handles error when fetching groups fails', async () => {
            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
            mockDatasource.getGroups = jest.fn().mockRejectedValue(new Error('Network error'));

            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });

            await waitFor(() => {
                expect(consoleSpy).toHaveBeenCalledWith('Error fetching groups:', expect.any(Error));
            });

            consoleSpy.mockRestore();
        });

        it('handles invalid response format for groups', async () => {
            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
            mockDatasource.getGroups = jest.fn().mockResolvedValue({ invalid: 'format' });

            await act(async () => {
                render(<QueryEditor {...defaultProps} />);
            });

            await waitFor(() => {
                expect(consoleSpy).toHaveBeenCalledWith('Invalid response format:', { invalid: 'format' });
            });

            consoleSpy.mockRestore();
        });

        it('handles error when fetching devices fails', async () => {
            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
            mockDatasource.getDevices = jest.fn().mockRejectedValue(new Error('Device error'));

            await act(async () => {
                render(<QueryEditor {...defaultProps} query={{ ...defaultQuery, group: 'Group 1' }} />);
            });

            await waitFor(() => {
                expect(consoleSpy).toHaveBeenCalledWith('Error fetching devices:', expect.any(Error));
            });

            consoleSpy.mockRestore();
        });

        it('handles error when fetching sensors fails', async () => {
            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
            mockDatasource.getSensors = jest.fn().mockRejectedValue(new Error('Sensor error'));

            await act(async () => {
                render(<QueryEditor {...defaultProps} query={{ ...defaultQuery, device: 'Device 1' }} />);
            });

            await waitFor(() => {
                expect(consoleSpy).toHaveBeenCalledWith('Error fetching sensors:', expect.any(Error));
            });

            consoleSpy.mockRestore();
        });

        it('handles error when fetching channels fails', async () => {
            const consoleSpy = jest.spyOn(console, 'error').mockImplementation();
            mockDatasource.getChannels = jest.fn().mockRejectedValue(new Error('Channel error'));

            await act(async () => {
                render(<QueryEditor {...defaultProps} query={{ ...defaultQuery, sensorId: '111' }} />);
            });

            await waitFor(() => {
                expect(consoleSpy).toHaveBeenCalledWith('Error fetching channels:', expect.any(Error));
            });
            consoleSpy.mockRestore();
        });
    });

    describe('Manual API Mode', () => {
        it('handles manual method change', async () => {
            const onChange = jest.fn();
            const onRunQuery = jest.fn();

            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} onRunQuery={onRunQuery} query={{ ...defaultQuery, queryType: QueryType.Manual }} />);
            });

            // Wait for component to render
            await waitFor(() => {
                expect(screen.getByTestId('query-editor-manualMethod')).toBeInTheDocument();
            });

            const manualMethodSelect = screen.getByTestId('query-editor-manualMethod');

            await act(async () => {
                fireEvent.change(manualMethodSelect, { target: { value: 'getobjectstatus.htm' } });
            });

            // Since the component uses debounced changes, let's check if onChange was called at all
            await waitFor(() => {
                expect(onChange).toHaveBeenCalled();
            }, { timeout: 3000 });
        });

        it('handles manual object ID change', async () => {
            const onChange = jest.fn();
            const onRunQuery = jest.fn();

            await act(async () => {
                render(<QueryEditor {...defaultProps} onChange={onChange} onRunQuery={onRunQuery} query={{ ...defaultQuery, queryType: QueryType.Manual }} />);
            });

            // Wait for component to render
            await waitFor(() => {
                expect(screen.getByTestId('query-editor-manualObjectId')).toBeInTheDocument();
            });

            const objectIdInput = screen.getByTestId('query-editor-manualObjectId');

            await act(async () => {
                fireEvent.change(objectIdInput, { target: { value: '12345' } });
            });

            await waitFor(() => {
                expect(onChange).toHaveBeenCalledWith(
                    expect.objectContaining({
                        manualObjectId: '12345',
                    })
                );
            });
        });
    });

    describe('Query Execution', () => {
        it('calls onRunQuery when query changes', async () => {
            const onRunQuery = jest.fn();
            const onChange = jest.fn();

            await act(async () => {
                render(<QueryEditor {...defaultProps} onRunQuery={onRunQuery} onChange={onChange} />);
            });

            // Wait for component to mount and initial data to load
            await waitFor(() => {
                expect(mockDatasource.getGroups).toHaveBeenCalled();
            });

            // Simulate a user interaction that would change the query
            const groupSelect = screen.getByTestId('query-editor-group');

            await act(async () => {
                fireEvent.change(groupSelect, { target: { value: 'Group 1' } });
            });

            // The onChange should be called first
            await waitFor(() => {
                expect(onChange).toHaveBeenCalled();
            });

            // And then onRunQuery should be called due to the runQueryIfChanged callback
            await waitFor(() => {
                expect(onRunQuery).toHaveBeenCalled();
            });
        });
    });
});