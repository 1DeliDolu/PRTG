import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export enum QueryType {
  Metrics = 'metrics',
  Raw = 'raw',
  Text = 'text',
  Manual = 'manual',
}

export interface MyQuery extends DataQuery {
  queryType: QueryType;
  group: string;
  groupId: string;
  device: string;
  deviceId: string;
  sensor: string;
  sensorId: string;
  channel: string;
  channelArray: string[];
  manualMethod?: string;
  manualObjectId?: string;
  property?: string;
  filterProperty?: string;
  includeGroupName?: boolean;
  includeDeviceName?: boolean;
  includeSensorName?: boolean;
  refId: string;

  // Add the streaming config
  streaming?: StreamingConfig;
  
  // Fields below can stay for backward compatibility but
  // recommend using the streaming object above for new code
  isStreaming?: boolean;
  streamInterval?: number;
  streamId?: string;
  panelId?: string | number;
  queryId?: string;
  cacheTime?: number;
  bufferSize?: number;
  updateMode?: 'full' | 'append';
}

// Organize streaming options better for clarity
export interface StreamingConfig {
  isStreaming?: boolean;      // Whether streaming is enabled
  streamInterval?: number;    // How often to fetch data (milliseconds)
  streamId?: string;          // Unique ID for this stream
  cacheTime?: number;         // How long to cache data (milliseconds)
  bufferSize?: number;        // Max points to buffer in memory
  updateMode?: 'full' | 'append'; // How to update the graph
}

// Add better defaults
export const DEFAULT_STREAMING_OPTIONS: StreamingOptions = {
  bufferSize: 100,
  updateMode: 'append',
  cacheTime: 6000, // 6 seconds default
  maxStreamsPerPanel: 10,
  defaultInterval: 1000, // 1 second default
};
export interface StreamingOptions {
  bufferSize: number;
  updateMode: 'full' | 'append';
  cacheTime: number;
  maxStreamsPerPanel: number;
  defaultInterval: number;
}

// Add new interface for manual API methods
export interface ManualApiMethod {
  label: string;
  value: string;
  description: string;
}

export const manualApiMethods: ManualApiMethod[] = [
 /*  {
    label: 'Get Object Property',
    value: 'getobjectproperty.htm',
    description: 'Retrieve specific property of an object',
  }, */
  {
    label: 'Get Sensor Details',
    value: 'getsensordetails.json',
    description: 'Get detailed information about a sensor',
  },
  {
    label: 'Get Status',
    value: 'getstatus.htm',
    description: 'Retrieve system status information',
  },
];

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  path?: string;
  cacheTime?: number;
  timeZone?: string;
}

export interface MySecureJsonData {
  apiKey?: string;
}


/* ################################### QUERY TYPE OPTION ###################################### */
export interface QueryTypeOptions {
  label: string;
  value: QueryType;
}

export const queryTypeOptions = Object.keys(QueryType).map((key) => ({
  label: key,
  value: QueryType[key as keyof typeof QueryType],
}));

export interface ListItem {
  name: string;
  visible_name: string;
}

/* ################################### PRTG ITEMS ###################################### */
export interface PRTGItem {
  active: boolean;
  active_raw: number;
  channel: string;
  channel_raw: string;
  datetime: string;
  datetime_raw: number;
  device: string;
  device_raw: string;
  group: string;
  group_raw: string;
  message: string;
  message_raw: string;
  objid: number;
  objid_raw: number;
  priority: string;
  priority_raw: number;
  sensor: string;
  sensor_raw: string;
  status: string;
  status_raw: number;
  tags: string;
  tags_raw: string;
  
}

export interface PRTGGroupListResponse {
  prtgversion: string;
  treesize: number;
  groups: PRTGItem[];
}

export interface PRTGGroupResponse {
  groups: PRTGItem[];
}

export interface PRTGDeviceListResponse {
  prtgversion: string;
  treesize: number;
  devices: PRTGItem[];
}

export interface PRTGDeviceResponse {
  devices: PRTGItem[];
}

export interface PRTGSensorListResponse {
  prtgversion: string;
  treesize: number;
  sensors: PRTGItem[];
}

export interface PRTGSensorResponse {
  sensors: PRTGItem[];
}

export interface PRTGChannelListResponse {
  prtgversion: string;
  treesize: number;
  values: PRTGItemChannel[];
}

export interface PRTGItemChannel {
  [key: string]: number | string;
  datetime: string;
}


/* ################################### Propert an filter prpoerty ################################################## */

export const filterPropertyList = [
  { name: 'active', visible_name: 'Active' },
  { name: 'message_raw', visible_name: 'Message' },
  { name: 'priority', visible_name: 'Priority' },
  { name: 'status', visible_name: 'Status' },
  { name: 'tags', visible_name: 'Tags' },
] as const;

export type FilterPropertyItem = typeof filterPropertyList[number];

export interface FilterPropertyOption {
  label: string;
  value: FilterPropertyItem['name'];
}

export const propertyList = [
  { name: 'group', visible_name: 'Group' },
  { name: 'device', visible_name: 'Device' },
  { name: 'sensor', visible_name: 'Sensor' },
] as const;

export type PropertyItem = typeof propertyList[number];

export interface PropertyOption {
  label: string;
  value: PropertyItem['name'];
}


/* ################################################## Query Selbst ################################################### */



export interface AnnotationsQuery extends MyQuery {
  from?: number;           // epoch datetime in milliseconds
  to?: number;            // epoch datetime in milliseconds
  limit?: number;         // default 100
  alertId?: number;
  dashboardId?: number;
  dashboardUID?: string;  // takes precedence over dashboardId
  panelId?: number;
  userId?: number;
  type?: 'alert' | 'annotation';
  tags?: string[];        // multiple tags for AND filtering
}

export interface Annotation {
  id?: number;
  alertId?: number;
  dashboardId?: number;
  dashboardUID?: string;
  panelId?: number | string;
  userId?: number;
  time: number;
  timeEnd?: number;
  title: string;
  text: string;
  tags?: string[];
  type?: 'alert' | 'annotation';
  data?: Record<string, any>;
}

export interface AnnotationResponse {
  annotations: Annotation[];
  total: number;
}