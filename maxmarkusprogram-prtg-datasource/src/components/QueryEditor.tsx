import React, { useEffect, useState, useMemo, useCallback, ChangeEvent } from 'react';
import {
  InlineField,
  Select,
  Stack,
  FieldSet,
  InlineSwitch,
  Input,
  AsyncMultiSelect,
} from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data'
import { DataSource } from '../datasource'
import { MyDataSourceOptions, MyQuery, queryTypeOptions, QueryType, propertyList, filterPropertyList, manualApiMethods } from '../types'

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {


  const isMetricsMode = query.queryType === QueryType.Metrics
  const isRawMode = query.queryType === QueryType.Raw
  const isTextMode = query.queryType === QueryType.Text
  const isManualMode = query.queryType === QueryType.Manual

  /* ===================================================== HOOKS ============================================================*/
  const [group, setGroup] = useState<string>(query.group || '')
  const [device, setDevice] = useState<string>(query.device || '')
  const [sensor, setSensor] = useState<string>(query.sensor || '')
  //@ts-ignore
  const [channel, setChannel] = useState<string>(query.channel || '')
  const [channelQuery, setChannelQuery] = useState<string[]>(query.channelArray || [])
  const [sensorId, setSensorId] = useState<string>(query.sensorId || '')
  const [manualMethod, setManualMethod] = useState<string>(query.manualMethod || '');
  const [manualObjectId, setManualObjectId] = useState<string>(query.manualObjectId || '');
  const [isStreaming, setIsStreaming] = useState<boolean>(query.isStreaming || false);
  const [streamInterval, setStreamInterval] = useState<number>(query.streamInterval || 1000);

  const [lists, setLists] = useState({
    groups: [] as Array<SelectableValue<string>>,
    devices: [] as Array<SelectableValue<string>>,
    sensors: [] as Array<SelectableValue<string>>,
    channels: [] as Array<SelectableValue<string>>,
    values: [] as Array<SelectableValue<string>>,
    properties: [] as Array<SelectableValue<string>>,
    filterProperties: [] as Array<SelectableValue<string>>,
  })

  const [isLoading, setIsLoading] = useState(false)


  /* ================================================== SORT ================================================== */
  lists.groups.sort((a, b) => (a.label ?? '').localeCompare(b.label ?? ''))
  lists.devices.sort((a, b) => (a.label ?? '').localeCompare(b.label ?? ''))
  lists.sensors.sort((a, b) => (a.label ?? '').localeCompare(b.label ?? ''))




  /* ================================================== FETCH GROUPS ================================================== */
  useEffect(() => {
    async function fetchGroups() {
      setIsLoading(true)
      try {
        const response = await datasource.getGroups()
        if (response && Array.isArray(response.groups)) {
          const groupOptions = response.groups.map((group) => ({
            label: group.group,
            value: group.group.toString(),
          }))
          setLists((prev) => ({
            ...prev,
            groups: groupOptions,
          }))
        } else {
          console.error('Invalid response format:', response)
        }
      } catch (error) {
        console.error('Error fetching groups:', error)
      }
      setIsLoading(false)
    }
    fetchGroups()
  }, [datasource])

  /* ================================================== FETCH DEVICES ================================================== */
  useEffect(() => {
    async function fetchDevices() {
      if (!group) {return};
      
      setIsLoading(true)
      try {
        const response = await datasource.getDevices(group)
        if (response && Array.isArray(response.devices)) {
          const filteredDevices = group ? response.devices.filter((device) => device.group === group) : response.devices

          const deviceOptions = filteredDevices.map((device) => ({
            label: device.device,
            value: device.device.toString(),
          }))
          setLists((prev) => ({
            ...prev,
            devices: deviceOptions,
          }))
        }
      } catch (error) {
        console.error('Error fetching devices:', error)
      }
      setIsLoading(false)
    }
    fetchDevices()
  }, [datasource, group])

  /* ================================================== FETCH SENSOR ================================================== */
  useEffect(() => {
    async function fetchSensors() {
      if (!device) {return};

      setIsLoading(true)
      try {
        const response = await datasource.getSensors(device)
        if (response && Array.isArray(response.sensors)) {
          const filteredSensors = device
            ? response.sensors.filter((sensor) => sensor.device === device)
            : response.sensors
          const sensorOptions = filteredSensors.map((sensor) => ({
            label: sensor.sensor,
            value: sensor.sensor.toString(),
          }))
          setLists((prev) => ({
            ...prev,
            sensors: sensorOptions,
          }))
        }
      } catch (error) {
        console.error('Error fetching sensors:', error)
      }
      setIsLoading(false)
    }
    fetchSensors()
  }, [datasource, device])

  /* ==================================================  FETCH CHANNEL ==================================================   */
  useEffect(() => {
    async function fetchChannels() {
        if (!sensorId) {
            return;
        }

        setIsLoading(true);
        try {
            const response = await datasource.getChannels(sensorId);
            if (!response) {
                console.error('Empty response received');
                return;
            }

            const channelData = response.values[0] || {};

            const channelOptions = Object.entries(channelData)
                .filter(([key]) => key !== 'datetime')
                .map(([key]) => ({
                    label: key,
                    value: key,
                }));

            setLists((prev) => ({
                ...prev,
                channels: channelOptions,
            }));

            if (query.channel && channelOptions.some(opt => opt.value === query.channel)) {
                setChannel(query.channel);
            }

        } catch (error) {
            console.error('Error fetching channels:', error);
        }
        setIsLoading(false);
    }

    fetchChannels();
}, [datasource, sensorId, query.channel]);

  useEffect(() => {
    if (isTextMode || isRawMode) {
      const propertyOptions = propertyList.map((item) => ({
        label: item.visible_name,
        value: item.name,
      }));

      // Filter property options
      const filterPropertyOptions = filterPropertyList.map((item) => ({
        label: item.visible_name,
        value: item.name,
      }));

      setLists((prev) => ({
        ...prev,
        properties: propertyOptions,
        filterProperties: filterPropertyOptions,
      }));
    }
  }, [isTextMode, isRawMode]);

  /* ==================================================  INITIAL VALUES  ================================================== */
 useEffect(() => {
    setGroup((prev) => query.group ?? prev);
    setDevice((prev) => query.device ?? prev); 
    setSensor((prev) => query.sensor ?? prev);
    setChannel((prev) => query.channel ?? prev);
    setSensorId((prev) => query.sensorId ?? prev);
    setManualMethod((prev) => query.manualMethod ?? prev);
    setManualObjectId((prev) => query.manualObjectId ?? prev);
  }, [query]);
 

  /* ==================================================  FIND IDs ================================================= */
  const findGroupId = useCallback(async (groupName: string) => {
    try {
      const response = await datasource.getGroups()
      if (response && Array.isArray(response.groups)) {
        const group = response.groups.find((g) => g.group === groupName)
        if (group) {
          return group.objid.toString()
        }
      }
    } catch (error) {
      console.error('Error finding group ID:', error)
    }
    return ''
  }, [datasource])

  const findDeviceId = useCallback(async (deviceName: string) => {
    try {
      const response = await datasource.getDevices(group)
      if (response && Array.isArray(response.devices)) {
        const device = response.devices.find((d) => d.device === deviceName)
        if (device) {
          return device.objid.toString()
        }
      }
    } catch (error) {
      console.error('Error finding device ID:', error)
    }
    return ''
  }, [datasource, group])

  const findSensorObjid = useCallback(async (sensorName: string) => {
    try {
      const response = await datasource.getSensors(device)
      if (response && Array.isArray(response.sensors)) {
        const sensor = response.sensors.find((s) => s.sensor === sensorName)
        if (sensor) {
          setSensorId(sensor.objid.toString())
          return sensor.objid.toString()
        } else {
          console.error('Sensor not found:', sensorName)
        }
      } else {
        console.error('Invalid response format:', response)
      }
    } catch (error) {
      console.error('Error fetching sensors:', error)
    }
    return ''
  }, [datasource, device, setSensorId])
  /* ==================================================  USE MEMO  ==================================================  */

  const groupOptions = useMemo(() => lists.groups, [lists.groups]);
  const deviceOptions = useMemo(() => lists.devices, [lists.devices]);
  const sensorOptions = useMemo(() => lists.sensors, [lists.sensors]);

  // Add new loadChannelOptions function
  const loadChannelOptions = async () => {
    if (!sensorId) {
      return [];
    }

    try {
      // eslint-disable-next-line no-console
      console.debug('Fetching channels for sensorId:', sensorId);
      const response = await datasource.getChannels(sensorId);
      
      // Debug logging
      // eslint-disable-next-line no-console
      console.debug('Channel response:', response);

      if (!response || !response.values || !response.values.length) {
        console.error('Invalid channel response:', response);
        return [];
      }

      const channelData = response.values[0];
      const options = Object.keys(channelData)
        .filter(key => key !== 'datetime')
        .map(key => ({
          label: key,
          value: key,
        }));

      // eslint-disable-next-line no-console
      console.debug('Processed channel options:', options);
      return options;

    } catch (error) {
      console.error('Error loading channels:', error);
      return [];
    }
  };
  /* ==================================================  EVENT HANDLERS ==================================================  */

    /* ==================================================  QUERY  ==================================================  */

  /* ==================================================  ONQUERYTYPESCHANGE ==================================================  */
  const onQueryTypeChange = useCallback((value: SelectableValue<QueryType>) => {
    onChange({ ...query, queryType: value.value! });
    onRunQuery();
  }, [query, onChange, onRunQuery]);

  /* ==================================================  ONGROUPCHANGE ==================================================  */
  const onGroupChange = useCallback(async (value: SelectableValue<string>) => {
    const groupObjId = await findGroupId(value.value!)
    setGroup(value.value!);
    onChange({
      ...query,
      group: value.value!,
      groupId: groupObjId,
    });
    setLists(prev => ({ ...prev, devices: [], sensors: [], channels: [] }));
    //onRunQuery();
  }, [query, onChange,/*  onRunQuery,  */findGroupId]);


  /* ==================================================  ONDEVICECHANGE ================================================= */
  const onDeviceChange = useCallback(async (value: SelectableValue<string>) => {
    const deviceObjId = await findDeviceId(value.value!)
    
    setDevice(value.value!);
    onChange({
      ...query,
      device: value.value!,
      deviceId: deviceObjId,
    });
    setLists(prev => ({ ...prev, sensors: [], channels: [] }));
    //onRunQuery();
  }, [query, onChange, /*  onRunQuery,  */findDeviceId]);


/* ==================================================  ONSENSORCHANGE ==================================================  */
  const onSensorChange = useCallback(async (value: SelectableValue<string>) => {
    if (!value.value) {
      return;
    }

    const sensorObjId = await findSensorObjid(value.value);

    setSensor(value.value);
    setSensorId(sensorObjId);
    setLists(prev => ({ ...prev, channels: [] }));

    onChange({
      ...query,
      sensor: value.value,
      sensorId: sensorObjId,
    });
    
    //onRunQuery();
  }, [query, onChange, /*  onRunQuery,  */findSensorObjid]);

  /* ==================================================  ONCHANNELCHANGE ==================================================  */
  const onChannelChange = (values: Array<SelectableValue<string>>) => {
    const selectedChannels = values.map(v => v.value || '');
    
    onChange({
      ...query,
      channel: selectedChannels[0] || '',
      channelArray: selectedChannels,
    });
    
    setChannelQuery(selectedChannels);
    setChannel(selectedChannels[0] || '');
    onRunQuery();
  };

/* ==================================================  ONPROPERTYCHANGE ==================================================  */
const onPropertyChange = (value: SelectableValue<string>) => {
  if (!value?.value) {return};
  
  onChange({ 
    ...query, 
    property: value.value,
  });
  onRunQuery();
};

/* ==================================================  ON FILTER PROPERTY CHANGE ==================================================  */
const onFilterPropertyChange = (value: SelectableValue<string>) => {
  if (!value?.value) {return};
  
  onChange({ 
    ...query, 
    filterProperty: value.value 
  });
  onRunQuery();
};

/* ==================================================  ON INCLUDE GROUP NAME ==================================================  */
const onIncludeGroupName = (event: ChangeEvent<HTMLInputElement>) => {
  onChange({ ...query, includeGroupName: event.currentTarget.checked })
  onRunQuery()
}


/* ==================================================  ON INCLUDE DEVICE NAME ==================================================  */
const onIncludeDeviceName = (event: React.ChangeEvent<HTMLInputElement>) => {
  onChange({ ...query, includeDeviceName: event.currentTarget.checked })
  onRunQuery()
}


/* ==================================================  ON INCLUDE SENSOR NAME ==================================================  */
const onIncludeSensorName = (event: ChangeEvent<HTMLInputElement>) => {
  onChange({ ...query, includeSensorName: event.currentTarget.checked })
  onRunQuery()
}

/* ==================================================  ON MANUAL METHOD CHANGE ==================================================  */
const onManualMethodChange = (value: SelectableValue<string>) => {
  setManualMethod(value.value!);
  onChange({
    ...query,
    manualMethod: value.value,
  });
  onRunQuery();
};


/* ==================================================  ON MANUAL OBJECT ID CHANGE ==================================================  */
const onManualObjectIdChange = (event: ChangeEvent<HTMLInputElement>) => {
  const value = event.currentTarget.value;
  setManualObjectId(value);
  onChange({
    ...query,
    manualObjectId: value,
  });
  onRunQuery();
};

/* ==================================================  ON STREAMING CHANGE ==================================================  */
const onStreamingChange = (event: ChangeEvent<HTMLInputElement>) => {
  const value = event.currentTarget.checked;
  setIsStreaming(value);
  onChange({ ...query, isStreaming: value });
  onRunQuery();
};

/* ==================================================  ON STREAM INTERVAL CHANGE ==================================================  */
const onStreamIntervalChange = (e: React.ChangeEvent<HTMLInputElement>) => {
  const value = parseInt(e.currentTarget.value, 10);
  setStreamInterval(value);
  onChange({ ...query, streamInterval: value });
  onRunQuery();
};

/* ================================================== DESTRUCTURING ================================================== */


/* ================================================== RENDER ================================================== */
return (
  <Stack direction="column" gap={2}>
    <Stack direction="row" gap={2}>
      <Stack direction="column" gap={1}>
        <InlineField label="Query Type" labelWidth={20} grow>
          <Select
            options={queryTypeOptions}
            value={query.queryType}
            onChange={onQueryTypeChange}
            width={47}
          />
        </InlineField>

        <InlineField label="Group" labelWidth={20} grow>
          <Select
            isLoading={isLoading}
            options={groupOptions}
            value={group}
            onChange={onGroupChange}
            width={47}
            allowCustomValue
            isClearable
            isDisabled={!query.queryType}
            placeholder="Select Group or type '*'"
          />
        </InlineField>

        <InlineField label="Device" labelWidth={20} grow>
          <Select
            isLoading={!lists.devices.length}
            options={deviceOptions}
            value={device}
            onChange={onDeviceChange}
            width={47}
            allowCustomValue
            isClearable
            isDisabled={!query.group}
            placeholder="Select Device or type '*'"
          />
        </InlineField>
      </Stack>

      <Stack direction="column" gap={1}>
        <InlineField label="Sensor" labelWidth={20} grow>
          <Select
            isLoading={!lists.sensors.length}
            options={sensorOptions}
            value={sensor}
            onChange={onSensorChange}
            width={47}
            allowCustomValue
            placeholder="Select Sensor or type '*'"
            isClearable
            isDisabled={!query.device}
          />
        </InlineField>
        <InlineField label="Channel" labelWidth={20} grow>
          <AsyncMultiSelect
            key={sensorId}
            loadOptions={loadChannelOptions}
            defaultOptions={true}
            value={(channelQuery || []).map(c => ({ label: c, value: c }))}
            onChange={onChannelChange}
            width={47}
            placeholder={sensorId ? "Select Channel" : "First select a sensor"}
            isClearable
            isDisabled={!sensorId}
          />
        </InlineField>
      </Stack>
    </Stack>
   

    {/*options for Metrics    */}
    {isMetricsMode && (
      <FieldSet label="Options">
        <Stack direction="row" gap={1}>
          <InlineField label="Include Group" labelWidth={16}>
            <InlineSwitch value={query.includeGroupName || false} onChange={onIncludeGroupName} />
          </InlineField>
          <InlineField label="Include Device" labelWidth={15}>
            <InlineSwitch value={query.includeDeviceName || false} onChange={onIncludeDeviceName} />
          </InlineField>
          <InlineField label="Include Sensor" labelWidth={15}>
            <InlineSwitch value={query.includeSensorName || false} onChange={onIncludeSensorName} />
          </InlineField>
          <InlineField label="Enable Streaming" labelWidth={16}>
            <InlineSwitch value={isStreaming} onChange={onStreamingChange} />
          </InlineField>
          {isStreaming && (
            <InlineField label="Stream Interval (ms)" labelWidth={20}>
              <Input
                type="number"
                value={streamInterval}
                onChange={onStreamIntervalChange}
                min={100}
                max={60000}
              />
            </InlineField>
          )}
        </Stack>
      </FieldSet>
    )}

    {/* Options for Text and Raw modes */}
    {(isTextMode || isRawMode) && (
      <FieldSet label="Options">
        <Stack direction="row" gap={2}>
          <InlineField label="Property" labelWidth={16} tooltip="Select property type">
            <Select
              options={lists.properties}
              value={query.property}
              onChange={onPropertyChange}
              width={32}
              placeholder="Select property"
              isClearable={false}
            />
          </InlineField>
          <InlineField label="Filter Property" labelWidth={16} tooltip="Select filter property">
            <Select
              options={lists.filterProperties}
              value={query.filterProperty}
              onChange={onFilterPropertyChange}
              width={32}
              placeholder="Select filter"
              isClearable={false}
            />
          </InlineField>
        </Stack>
      </FieldSet>
    )}

    {/* Manual API Query Section */}
    {isManualMode && (
      <FieldSet label="Manual API Query">
        <Stack direction="row" gap={2}>
          <InlineField label="API Method" labelWidth={16} tooltip="Select or enter a custom PRTG API endpoint">
            <Select
              options={manualApiMethods}
              value={manualMethod}
              onChange={onManualMethodChange}
              width={32}
              placeholder="Select or enter API method"
              allowCustomValue={true}
              onCreateOption={(customValue) => {
                setManualMethod(customValue);
                onChange({
                  ...query,
                  manualMethod: customValue,
                });
                onRunQuery();
              }}
              isClearable
            />
          </InlineField>
          <InlineField label="Object ID" labelWidth={16} tooltip="Object ID from selected sensor">
            <Input
              value={manualObjectId || sensorId}
              onChange={onManualObjectIdChange}
              placeholder="Automatically filled from sensor"
              width={32}
              type="text"
              disabled={!!sensorId}
            />
          </InlineField>
        </Stack>
      </FieldSet>
    )}

  </Stack>
)
}

