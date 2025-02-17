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

  async getDevices(): Promise<PRTGDeviceListResponse> {
    return this.getResource('devices')
  }

  async getSensors(): Promise<PRTGSensorListResponse> {
    return this.getResource('sensors')
  }

  async getChannels(objid: string): Promise<PRTGChannelListResponse> {
    if (!objid) {
      throw new Error('objid is required')
    }
    return this.getResource(`channels/${objid}`)
  }
  annotations?: AnnotationSupport<MyQuery, AnnotationQuery<MyQuery>> | undefined
}
