import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export enum QueryType {
  Metrics = 'metrics',
  Raw = 'raw',
  Text = 'text'
}

export interface MyQuery extends DataQuery {
  group: string;
  device: string;
  sensor: string;
  objid: number | string;
  channel: string;
  queryType: QueryType;
  property: string;
  filterProperty: string;
  includeGroupName: boolean;
  includeDeviceName: boolean;
  includeSensorName: boolean;
  groups: Array<string>;
  devices: Array<string>;
  sensors: Array<string>;
  channels: Array<string>;
}

export interface DataPoint {
  Time: number;
  Value: number | string;
}

export interface DataSourceResponse {
  datapoints: DataPoint[];
}

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  path?: string;
  cacheTime?: number;
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





