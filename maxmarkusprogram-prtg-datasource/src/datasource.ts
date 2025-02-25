import { DataSourceInstanceSettings, ScopedVars } from '@grafana/data'
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
    super(instanceSettings);
    this.annotations = {
      QueryEditor: null, 
      processEvents: null, 
    };
  }

  /* =================================== APPLYTEMPLATEVARIABLES ====================================== */
  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars) {
    return {
      ...query,
      channel: getTemplateSrv().replace(query.channel, scopedVars),
    }
  }

  filterQuery(query: MyQuery): boolean {
    return !!query.channel
  }

  /* =================================== GETRESOURCE ====================================== */
  async getGroups(): Promise<PRTGGroupListResponse> {
    return this.getResource('groups')
  }

  async getDevices(group: string): Promise<PRTGDeviceListResponse> {
    if (!group) {
      throw new Error('group is required')
    }
    return this.getResource(`devices/${encodeURIComponent(group)}`)
  }

  async getSensors(device: string): Promise<PRTGSensorListResponse> {
    if (!device) {
      throw new Error('device is required');
    }
    return this.getResource(`sensors/${encodeURIComponent(device)}`);
  }

  async getChannels(sensorId: string): Promise<PRTGChannelListResponse> {
    if (!sensorId) {
      throw new Error('sensorId is required')
    }
    return this.getResource(`channels/${encodeURIComponent(sensorId)}`)
  }

  /* =================================== ANNOTATIONS ====================================== */
  annotations: {
  }
}
