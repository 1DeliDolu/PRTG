import { 
  DataSourceInstanceSettings, 
  ScopedVars, 
  AnnotationEvent,
  DataFrame,

} from '@grafana/data';
import { 
  DataSourceWithBackend, 
  getTemplateSrv,

} from '@grafana/runtime';
import { Observable, from, merge, throwError } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import {
  MyQuery,
  MyDataSourceOptions,
  PRTGGroupListResponse,
  PRTGDeviceListResponse,
  PRTGSensorListResponse,
  PRTGChannelListResponse,
  QueryType,
} from './types'

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {

  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars) {
    const replaced = getTemplateSrv().replace(query.channel, scopedVars);
    return {
      ...query,
      channel: replaced,
    }
  }

  filterQuery(query: MyQuery): boolean {
    return !!query.channel
  }

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
      throw new Error('sensorId is required');
    }
    return this.getResource(`channels/${encodeURIComponent(sensorId)}`);
  }

  annotations = {
    QueryEditor: undefined,
    processEvents: (anno: any, data: DataFrame[]): Observable<AnnotationEvent[]> => {
      const events: AnnotationEvent[] = [];
      
      data.forEach((frame) => {
        const timeField = frame.fields.find((field) => field.name === 'Time');
        const valueField = frame.fields.find((field) => field.name === 'Value');
        
        if (timeField && valueField) {
          const firstTime = timeField.values[0];
          const lastTime = timeField.values[timeField.values.length - 1];
          const firstValue = valueField.values[0];
          const panelId = typeof anno.panelId === 'number' ? anno.panelId : undefined;
          const source = frame.name || 'PRTG Channel';

          events.push({
            time: firstTime,
            timeEnd: lastTime !== firstTime ? lastTime : undefined,
            title: source,
            text: `Value: ${firstValue}`,
            tags: ['prtg', `value:${firstValue}`, `source:${source}`],
            panelId: panelId
          });
        }
      });

      return from([events]);
    },
  };

}
