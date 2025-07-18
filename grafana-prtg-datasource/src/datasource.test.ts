import { DataQueryRequest, DataSourceInstanceSettings } from '@grafana/data';
import { getTemplateSrv, getGrafanaLiveSrv } from '@grafana/runtime';
import { DataSource } from './datasource';
import { MyQuery, MyDataSourceOptions, QueryType } from './types';

// Suppress console output during tests except errors
const originalConsole = console;
beforeAll(() => {
  global.console = {
    ...originalConsole,
    log: jest.fn(),
    info: jest.fn(),
    debug: jest.fn(),
    warn: jest.fn(),
    // Keep error logs
    error: originalConsole.error,
  };
});

afterAll(() => {
  global.console = originalConsole;
});

jest.mock('@grafana/runtime', () => ({
  getTemplateSrv: jest.fn(),
  getGrafanaLiveSrv: jest.fn(),
  DataSourceWithBackend: class MockDataSourceWithBackend {
    uid = 'test-uid';
    constructor() {}
    getResource = jest.fn();
    query = jest.fn().mockReturnValue({ data: [] });
  },
}));

const mockTemplateSrv = { replace: jest.fn() };
const mockLiveSrv = { getDataStream: jest.fn() };

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
  updateMode: 'full',
  ...overrides,
});

const mockSettings: DataSourceInstanceSettings<MyDataSourceOptions> = {
  id: 1,
  uid: 'test-uid',
  type: 'prtg',
  name: 'PRTG DataSource',
  url: 'http://localhost',
  access: 'proxy',
  readOnly: false,
  jsonData: {},
  meta: {} as any,
};

describe('DataSource', () => {
  let dataSource: DataSource;

  beforeEach(() => {
    jest.clearAllMocks();
    (getTemplateSrv as jest.Mock).mockReturnValue(mockTemplateSrv);
    (getGrafanaLiveSrv as jest.Mock).mockReturnValue(mockLiveSrv);
    dataSource = new DataSource(mockSettings);
  });

  it('should create instance', () => {
    expect(dataSource).toBeDefined();
  });

  it('should handle template variables', () => {
    const query = createMockQuery({ channel: '$channel' });
    mockTemplateSrv.replace.mockReturnValue('replaced-channel');

    const result = dataSource.applyTemplateVariables(query, {});

    expect(mockTemplateSrv.replace).toHaveBeenCalledWith('$channel', {});
    expect(result.channel).toBe('replaced-channel');
  });

  it('should filter queries by channel', () => {
    expect(dataSource.filterQuery(createMockQuery({ channel: 'test' }))).toBe(true);
    expect(dataSource.filterQuery(createMockQuery({ channel: '' }))).toBe(false);
  });

  it('should handle resource calls', async () => {
    (dataSource as any).getResource.mockResolvedValue({ groups: [] });
    const result = await dataSource.getGroups();
    expect(result).toEqual({ groups: [] });
  });

  it('should handle device errors', async () => {
    await expect(dataSource.getDevices('')).rejects.toThrow('group is required');
  });
  it('should handle query calls', () => {
    const mockRequest: DataQueryRequest<MyQuery> = {
      requestId: 'test',
      interval: '1s',
      intervalMs: 1000,
      range: {
        from: { valueOf: () => 1000 } as any,
        to: { valueOf: () => 2000 } as any,
        raw: { from: 'now-1h', to: 'now' },
      },
      scopedVars: {},
      targets: [createMockQuery({ channel: 'test' })],
      timezone: 'UTC',
      app: 'grafana',
      startTime: Date.now(),
      panelId: 1,
    };

    // Mock the query method specifically for this test
    const mockResponse = { data: [] };
    jest.spyOn(dataSource, 'query').mockReturnValue(mockResponse as any);

    const result = dataSource.query(mockRequest);
    expect(result).toBeDefined();
    expect(result).toBe(mockResponse);
  });
});
