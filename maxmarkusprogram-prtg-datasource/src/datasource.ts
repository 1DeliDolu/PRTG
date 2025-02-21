import { AnnotationQuery, AnnotationSupport, DataSourceInstanceSettings, ScopedVars } from '@grafana/data'
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime'
import {
  MyQuery,
  MyDataSourceOptions,
  PRTGGroupListResponse,
  PRTGDeviceListResponse,
  PRTGSensorListResponse,
  PRTGChannelListResponse,
} from './types'

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings)
  }

  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars) {
    return {
      ...query,
      channel: getTemplateSrv().replace(query.channel, scopedVars),
    }
  }

  filterQuery(query: MyQuery): boolean {
    // if no query has been provided, prevent the query from being executed
    return !!query.channel
  }

  async getGroups(): Promise<PRTGGroupListResponse> {
    return this.getResource('groups')
  }

  async getDevices(group: string): Promise<PRTGDeviceListResponse> {
    if (!group) {
      throw new Error('group is required')
    }
    // Change this line to use path parameter instead of query parameter
    return this.getResource(`devices/${encodeURIComponent(group)}`)
  }

  async getSensors(device: string): Promise<PRTGSensorListResponse> {
    if (!device) {
      throw new Error('device is required');
    }
    // Change to use path parameter instead of query parameter
    return this.getResource(`sensors/${encodeURIComponent(device)}`);
  }

  async getChannels(objid: string): Promise<PRTGChannelListResponse> {
    if (!objid) {
      throw new Error('objid is required')
    }
    return this.getResource(`channels/${objid}`)
  }

  

  annotations?: AnnotationSupport<MyQuery, AnnotationQuery<MyQuery>> | undefined
}
