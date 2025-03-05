import { 
  DataSourceInstanceSettings, 
  ScopedVars, 
  AnnotationEvent,
  DataFrame,
  DataQueryRequest,
  DataQueryResponse,
  LiveChannelScope,
} from '@grafana/data';
import { 
  DataSourceWithBackend, 
  getTemplateSrv,
  getGrafanaLiveSrv,
} from '@grafana/runtime';
import { Observable, from, merge } from 'rxjs';
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
  }

  /* =================================== APPLYTEMPLATEVARIABLES ====================================== */
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
      throw new Error('sensorId is required');
    }
    // If you need logging, consider implementing a proper logging mechanism
    const response = await this.getResource(`channels/${encodeURIComponent(sensorId)}`);
    return response;
  }

  /* =================================== ANNOTATIONS ====================================== */
  annotations = {
    QueryEditor: undefined,
    processEvents: (anno: any, data: DataFrame[]): Observable<AnnotationEvent[]> => {
      const events: AnnotationEvent[] = [];
      
      // Use annotation query values if available
      const sourceQuery = anno.target || {};
      
      data.forEach((frame) => {
        const timeField = frame.fields.find((field) => field.name === 'Time');
        const valueField = frame.fields.find((field) => field.name === 'Value');

        
        if (timeField && valueField) {
          const firstTime = timeField.values[0];
          const lastTime = timeField.values[timeField.values.length - 1];
          const firstValue = valueField.values[0];
          const panelId = typeof anno.panelId === 'number' ? anno.panelId : undefined;

          // Use source from annotation query or default to frame name
          const source = sourceQuery.from || frame.name || 'PRTG Channel';

          const event: AnnotationEvent = {
            time: firstTime,
            timeEnd: lastTime !== firstTime ? lastTime : undefined,
            title: source,
            text: `Value: ${firstValue}`,
            tags: ['prtg', `value:${firstValue}`, `source:${source}`],
            panelId: panelId
          };

          events.push(event);
        }
      });

      return from([events]);
    },
  };

  query(request: DataQueryRequest<MyQuery>): Observable<DataQueryResponse> {
    // Handle streaming queries
    const observables = request.targets.map((query) => {
      if (query.isStreaming) {
        // Create a unique path for each streaming query based on PRTG parameters
        const streamPath = `prtg-stream/${query.sensorId}/${query.channelArray?.join('-')}/${query.streamInterval}`;
        
        return getGrafanaLiveSrv().getDataStream({
          addr: {
            scope: LiveChannelScope.DataSource,
            namespace: this.uid,
            path: streamPath,
            data: {
              ...query,
              sensorId: query.sensorId,
              channels: query.channelArray,
              interval: query.streamInterval,
              group: query.group,
              device: query.device,
              sensor: query.sensor,
            },
          },
        });
      }
      
      // For non-streaming queries, use the regular query handling
      return from(super.query(request));
    });

    return merge(...observables);
  }
}
