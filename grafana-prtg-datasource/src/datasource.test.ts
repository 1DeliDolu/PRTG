/// <reference types="jest" />
import { DataSource } from './datasource';
import { DataSourceInstanceSettings, DataFrame, DataQueryRequest, ScopedVars, FieldType, LiveChannelScope } from '@grafana/data';
import { getTemplateSrv, getGrafanaLiveSrv } from '@grafana/runtime';
import { of, throwError } from 'rxjs';
import { MyQuery, MyDataSourceOptions, QueryType } from './types';

// Mock Grafana runtime
jest.mock('@grafana/runtime', () => ({
    getTemplateSrv: jest.fn(),
    getGrafanaLiveSrv: jest.fn(),
    DataSourceWithBackend: class MockDataSourceWithBackend {
        constructor(instanceSettings: any) {
            this.instanceSettings = instanceSettings;
        }
        getResource = jest.fn();
        // Don't mock query here - let the child class override it
        uid = 'test-uid';
        instanceSettings: any;
    },
}));

const mockGetTemplateSrv = getTemplateSrv as jest.MockedFunction<typeof getTemplateSrv>;
const mockGetGrafanaLiveSrv = getGrafanaLiveSrv as jest.MockedFunction<typeof getGrafanaLiveSrv>;

describe('DataSource', () => {
    let dataSource: DataSource;
    let mockInstanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>;
    
    const createMockQuery = (overrides: Partial<MyQuery> = {}): MyQuery => ({
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
        includeGroupName: false,
        includeDeviceName: false,
        includeSensorName: false,
        isStreaming: false,
        streamInterval: 2500,
        ...overrides,
    });

    beforeEach(() => {
        mockInstanceSettings = {
            id: 1,
            uid: 'test-uid',
            type: 'prtg',
            name: 'PRTG Test',
            url: 'http://test.com',
            access: 'proxy',
            readOnly: false,
            jsonData: {},
            meta: {} as any,
        };

        dataSource = new DataSource(mockInstanceSettings);
        
        // Mock getResource method
        dataSource.getResource = jest.fn();
        // Mock the parent query method ONLY - don't mock the instance query method
        const parentProto = Object.getPrototypeOf(Object.getPrototypeOf(dataSource));
        parentProto.query = jest.fn().mockReturnValue(of({ data: [] }));
        // Mock console.error to prevent test failures from logging
        jest.spyOn(console, 'error').mockImplementation(() => {});
    });

    afterEach(() => {
        jest.clearAllMocks();
        jest.restoreAllMocks();
    });

    describe('constructor', () => {
        it('should create instance with settings', () => {
            expect(dataSource).toBeInstanceOf(DataSource);
        });
    });

    describe('applyTemplateVariables', () => {
        it('should replace template variables in channel', () => {
            const query = createMockQuery({ channel: '$channel_var' });
            const scopedVars: ScopedVars = {};
            const mockReplace = jest.fn().mockReturnValue('replaced_channel');
            
            mockGetTemplateSrv.mockReturnValue({ replace: mockReplace } as any);

            const result = dataSource.applyTemplateVariables(query, scopedVars);

            expect(mockReplace).toHaveBeenCalledWith('$channel_var', scopedVars);
            expect(result).toEqual({ ...query, channel: 'replaced_channel' });
        });

        it('should handle undefined channel', () => {
            const query = createMockQuery({ channel: undefined });
            const scopedVars: ScopedVars = {};
            const mockReplace = jest.fn().mockReturnValue(undefined);
            
            mockGetTemplateSrv.mockReturnValue({ replace: mockReplace } as any);

            const result = dataSource.applyTemplateVariables(query, scopedVars);

            expect(mockReplace).toHaveBeenCalledWith(undefined, scopedVars);
            expect(result).toEqual({ ...query, channel: undefined });
        });
    });

    describe('filterQuery', () => {
        it('should return true when query has channel', () => {
            const query = createMockQuery({ channel: 'test-channel' });
            expect(dataSource.filterQuery(query)).toBe(true);
        });

        it('should return false when query has no channel', () => {
            const query = createMockQuery({ channel: '' });
            expect(dataSource.filterQuery(query)).toBe(false);
        });

        it('should return false when channel is undefined', () => {
            const query = createMockQuery({ channel: undefined });
            expect(dataSource.filterQuery(query)).toBe(false);
        });
    });

    describe('getGroups', () => {
        it('should call getResource with groups endpoint', async () => {
            const mockResponse = { groups: [] };
            (dataSource.getResource as jest.Mock).mockResolvedValue(mockResponse);

            const result = await dataSource.getGroups();

            expect(dataSource.getResource).toHaveBeenCalledWith('groups');
            expect(result).toEqual(mockResponse);
        });
    });

    describe('getDevices', () => {
        it('should call getResource with encoded group parameter', async () => {
            const mockResponse = { devices: [] };
            (dataSource.getResource as jest.Mock).mockResolvedValue(mockResponse);

            const result = await dataSource.getDevices('test group');

            expect(dataSource.getResource).toHaveBeenCalledWith('devices/test%20group');
            expect(result).toEqual(mockResponse);
        });

        it('should throw error when group is not provided', async () => {
            await expect(dataSource.getDevices('')).rejects.toThrow('group is required');
        });
    });

    describe('getSensors', () => {
        it('should call getResource with encoded device parameter', async () => {
            const mockResponse = { sensors: [] };
            (dataSource.getResource as jest.Mock).mockResolvedValue(mockResponse);

            const result = await dataSource.getSensors('test device');

            expect(dataSource.getResource).toHaveBeenCalledWith('sensors/test%20device');
            expect(result).toEqual(mockResponse);
        });

        it('should throw error when device is not provided', async () => {
            await expect(dataSource.getSensors('')).rejects.toThrow('device is required');
        });
    });

    describe('getChannels', () => {
        it('should call getResource with encoded sensorId parameter', async () => {
            const mockResponse = { channels: [] };
            (dataSource.getResource as jest.Mock).mockResolvedValue(mockResponse);

            const result = await dataSource.getChannels('123');

            expect(dataSource.getResource).toHaveBeenCalledWith('channels/123');
            expect(result).toEqual(mockResponse);
        });

        it('should throw error when sensorId is not provided', async () => {
            await expect(dataSource.getChannels('')).rejects.toThrow('sensorId is required');
        });
    });

    describe('annotations.processEvents', () => {
        it('should process DataFrame and return annotation events', (done: jest.DoneCallback) => {
            const mockFrame: DataFrame = {
                name: 'Test Frame',
                fields: [
                    { name: 'Time', values: [1000, 2000], type: FieldType.time, config: {} },
                    { name: 'Value', values: [100, 200], type: FieldType.number, config: {} },
                ],
                length: 2,
            };

            const anno = { panelId: 1 };
            const data = [mockFrame];

            dataSource.annotations.processEvents(anno, data).subscribe(events => {
                expect(events).toHaveLength(1);
                expect(events[0]).toEqual({
                    time: 1000,
                    timeEnd: 2000,
                    title: 'Test Frame',
                    text: 'Value: 100',
                    tags: ['prtg', 'value:100', 'source:Test Frame'],
                    panelId: 1
                });
                done();
            });
        });

        it('should handle frame without Time or Value fields', (done: jest.DoneCallback) => {
            const mockFrame: DataFrame = {
                name: 'Test Frame',
                fields: [
                    { name: 'Other', values: [1, 2], type: FieldType.string, config: {} },
                ],
                length: 2,
            };

            const anno = {};
            const data = [mockFrame];

            dataSource.annotations.processEvents(anno, data).subscribe(events => {
                expect(events).toHaveLength(0);
                done();
            });
        });
    });

    describe('query', () => {
        let mockRequest: DataQueryRequest<MyQuery>;

        beforeEach(() => {
            mockRequest = {
                targets: [],
                range: {
                    from: { valueOf: () => 1000 } as any,
                    to: { valueOf: () => 2000 } as any,
                },
                panelId: 1,
            } as any;
        });

        it('should handle regular queries', (done: jest.DoneCallback) => {
            const regularQuery = createMockQuery({ channel: 'test' });
            mockRequest.targets = [regularQuery];

            // The parent query method is already mocked in beforeEach
            const mockQueryResponse = { data: [] };
            const parentProto = Object.getPrototypeOf(Object.getPrototypeOf(dataSource));
            (parentProto.query as jest.Mock).mockReturnValue(of(mockQueryResponse));

            dataSource.query(mockRequest).subscribe(response => {
                expect(response).toEqual(mockQueryResponse);
                done();
            });
        });

        it('should handle streaming queries', (done: jest.DoneCallback) => {
            const streamingQuery = createMockQuery({ 
                channel: 'test', 
                isStreaming: true, 
                queryType: QueryType.Metrics 
            });
            mockRequest.targets = [streamingQuery];

            const mockFrameData = [{ 
                meta: { 
                    someExistingProp: 'value'
                } 
            }];
            const mockLiveData = { data: mockFrameData };
            const mockStream = of(mockLiveData);
            
            const mockGetDataStream = jest.fn().mockReturnValue(mockStream);
            mockGetGrafanaLiveSrv.mockReturnValue({
                getDataStream: mockGetDataStream
            } as any);            dataSource.query(mockRequest).subscribe({
                next: (response) => {
                    try {
                        expect(mockGetDataStream).toHaveBeenCalled();
                        
                        // Verify the call arguments
                        const callArgs = mockGetDataStream.mock.calls[0][0];
                        expect(callArgs.addr.scope).toBe(LiveChannelScope.DataSource);
                        expect(callArgs.addr.namespace).toBe('test-uid');
                        expect(callArgs.addr.path).toMatch(/^prtg-stream\//);
                        
                        expect(response.data).toHaveLength(1);
                        expect(response.data[0].meta).toEqual(expect.objectContaining({
                            streaming: true,
                            preferredVisualisationType: 'graph'
                        }));
                        done();
                    } catch (error) {
                        done(error);
                    }
                },
                error: (err) => {
                    done(err);
                }
            });
        }, 10000);

        it('should return empty data when no targets', (done: jest.DoneCallback) => {
            mockRequest.targets = [];

            dataSource.query(mockRequest).subscribe(response => {
                expect(response).toEqual({ data: [] });
                done();
            });
        });

        it('should handle streaming errors', (done: jest.DoneCallback) => {
            const streamingQuery = createMockQuery({ 
                channel: 'test', 
                isStreaming: true, 
                queryType: QueryType.Metrics 
            });
            mockRequest.targets = [streamingQuery];

            const errorStream = throwError(() => new Error('Stream failed'));
            
            const mockGetDataStream = jest.fn().mockReturnValue(errorStream);
            mockGetGrafanaLiveSrv.mockReturnValue({
                getDataStream: mockGetDataStream
            } as any);

            dataSource.query(mockRequest).subscribe({
                next: () => {
                    // Should not reach here
                    done(new Error('Expected error but got success'));
                },
                error: (err) => {
                    try {
                        expect(mockGetDataStream).toHaveBeenCalled();
                        expect(err.message).toContain('Streaming error');
                        done();
                    } catch (error) {
                        done(error);
                    }
                }
            });
        }, 10000);

        it('should handle mixed streaming and regular queries', (done: jest.DoneCallback) => {
            const streamingQuery = createMockQuery({ 
                refId: 'A',
                channel: 'test-stream', 
                isStreaming: true, 
                queryType: QueryType.Metrics 
            });
            const regularQuery = createMockQuery({ 
                refId: 'B',
                channel: 'test-regular' 
            });
            mockRequest.targets = [streamingQuery, regularQuery];

            // Mock streaming
            const mockFrameData = [{ meta: { someExistingProp: 'value' } }];
            const mockLiveData = { data: mockFrameData };
            const mockStream = of(mockLiveData);
            const mockGetDataStream = jest.fn().mockReturnValue(mockStream);
            mockGetGrafanaLiveSrv.mockReturnValue({
                getDataStream: mockGetDataStream
            } as any);

            // Mock regular query
            const mockRegularResponse = { data: [{ refId: 'B' }] };
            const parentProto = Object.getPrototypeOf(Object.getPrototypeOf(dataSource));
            (parentProto.query as jest.Mock).mockReturnValue(of(mockRegularResponse));

            let responseCount = 0;
            const expectedResponses = 2;

            dataSource.query(mockRequest).subscribe({
                next: (response) => {
                    responseCount++;
                    try {
                        // Should receive both streaming and regular responses
                        expect(response.data).toBeDefined();
                        
                        if (responseCount === expectedResponses) {
                            expect(mockGetDataStream).toHaveBeenCalled();
                            done();
                        }
                    } catch (error) {
                        done(error);
                    }
                },
                error: (err) => {
                    done(err);
                }
            });
        }, 10000);
    });

    describe('getStreamId', () => {
        it('should generate stream ID from query components with channel', () => {
            const query = createMockQuery({
                refId: 'A',
                sensorId: '123',
                channel: 'test-channel',
                channelArray: [] // Empty array should fallback to channel
            });

            // The getStreamId method expects panelId to be added to the query object
            // This is done in the query method from request.panelId
            const queryWithPanelId = { ...query, panelId: '1' };
            const streamId = (dataSource as any).getStreamId(queryWithPanelId);
            expect(streamId).toBe('1_A_123_test-channel');
        });

        it('should generate stream ID with undefined channelArray', () => {
            const query = createMockQuery({
                refId: 'A',
                sensorId: '123',
                channel: 'test-channel',
                channelArray: undefined
            });

            const queryWithPanelId = { ...query, panelId: '1' };
            const streamId = (dataSource as any).getStreamId(queryWithPanelId);
            expect(streamId).toBe('1_A_123_test-channel');
        });

        it('should handle channelArray over channel', () => {
            const query = createMockQuery({
                refId: 'A',
                sensorId: '123',
                channel: 'should-be-ignored',
                channelArray: ['ch1', 'ch2']
            });

            const queryWithPanelId = { ...query, panelId: '1' };
            const streamId = (dataSource as any).getStreamId(queryWithPanelId);
            expect(streamId).toBe('1_A_123_ch1-ch2');
        });

        it('should use defaults for missing values', () => {
            const query = createMockQuery({ 
                refId: 'A',
                sensorId: '',
                channel: '',
                channelArray: []
            });

            const streamId = (dataSource as any).getStreamId(query);
            // Only non-empty values are included, so empty sensorId and channel are filtered out
            expect(streamId).toBe('default_A');
        });
    });

    describe('getStreamStatus', () => {
        it('should call getResource with stream status endpoint', async () => {
            const mockResponse = { status: 'active' };
            (dataSource.getResource as jest.Mock).mockResolvedValue(mockResponse);

            const result = await dataSource.getStreamStatus('test-stream');

            expect(dataSource.getResource).toHaveBeenCalledWith('stream-status/test-stream');
            expect(result).toEqual(mockResponse);
        });
    });

    describe('stopStream', () => {
        it('should call getResource with stop stream endpoint', async () => {
            (dataSource.getResource as jest.Mock).mockResolvedValue(undefined);

            await dataSource.stopStream('test-stream');

            expect(dataSource.getResource).toHaveBeenCalledWith('stop-stream/test-stream');
        });
    });
});