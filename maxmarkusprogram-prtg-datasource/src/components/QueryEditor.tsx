import React, { useEffect, useState, useMemo, useCallback, ChangeEvent, useRef } from 'react';
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
  const prevQueryRef = useRef<MyQuery | null>(null);

  const runQueryIfChanged = useCallback(() => {
    const currentQuery = JSON.stringify(query);
    const prevQuery = JSON.stringify(prevQueryRef.current);

    if (currentQuery !== prevQuery) {
      prevQueryRef.current = query;
      onRunQuery();
    }
  }, [query, onRunQuery]);

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
  lists.channels.sort((a, b) => (a.label ?? '').localeCompare(b.label ?? ''))




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
    // Add this line to restore channel selections
    setChannelQuery((prev) => query.channelArray || prev || []);
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

  // Add new memoized selected values
  const selectedGroup = useMemo(() => {
    return groupOptions.find(option => option.value === group) || (group ? {label: group, value: group} : null);
  }, [groupOptions, group]);

  const selectedDevice = useMemo(() => {
    return deviceOptions.find(option => option.value === device) || (device ? {label: device, value: device} : null);
  }, [deviceOptions, device]);

  const selectedSensor = useMemo(() => {
    return sensorOptions.find(option => option.value === sensor) || (sensor ? {label: sensor, value: sensor} : null);
  }, [sensorOptions, sensor]);

  // Add new loadChannelOptions function with useMemo
  const loadChannelOptions = useMemo(() => async () => {
    if (!sensorId) {
      return [];
    }

    try {
      const response = await datasource.getChannels(sensorId);

      if (!response) {
        console.warn('No response received from getChannels');
        return [];
      }

      // Check if response has the expected structure
      if (typeof response === 'object' && 'values' in response) {
        const values = response.values;
        if (!Array.isArray(values) || values.length === 0) {
          console.warn('No channel values found in response');
          return [];
        }

        const channelData = values[0];
        if (typeof channelData !== 'object') {
          console.warn('Invalid channel data format');
          return [];
        }

        return Object.keys(channelData)
          .filter(key => key !== 'datetime')
          .map(key => ({
            label: key,
            value: key,
          }));
      }

      console.warn('Unexpected response format:', response);
      return [];
    } catch (error: any) {
      console.error('Error loading channels:', error?.message || error);
      return [];
    }
  }, [sensorId, datasource]);
  /* ==================================================  EVENT HANDLERS ==================================================  */

    /* ==================================================  QUERY  ==================================================  */

  /* ==================================================  ONQUERYTYPESCHANGE ==================================================  */
  const onQueryTypeChange = useCallback((value: SelectableValue<QueryType>) => {
    onChange({ ...query, queryType: value.value! });
    runQueryIfChanged();
  }, [query, onChange, runQueryIfChanged]);

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
    runQueryIfChanged();
  }, [query, onChange, runQueryIfChanged,  findGroupId]);


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
    runQueryIfChanged();
  }, [query, onChange, runQueryIfChanged,  findDeviceId]);


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
    
    runQueryIfChanged();
  }, [query, onChange,  runQueryIfChanged,  findSensorObjid]);

  /* ==================================================  ONCHANNELCHANGE ==================================================  */
  const onChannelChange = useCallback((values: Array<SelectableValue<string>>) => {
    const selectedChannels = values.map(v => v.value!);
    
    // Update both local state and query model
    setChannelQuery(selectedChannels);
    
    // Ensure we store both value and label for each channel
    onChange({
      ...query,
      channel: selectedChannels[0] || '',
      channelArray: selectedChannels,
    });
    
    runQueryIfChanged();
  }, [query, onChange, runQueryIfChanged]);

/* ==================================================  ONPROPERTYCHANGE ==================================================  */
const onPropertyChange = (value: SelectableValue<string>) => {
  if (!value?.value) {return};
  
  onChange({ 
    ...query, 
    property: value.value,
  });
  runQueryIfChanged();
};

/* ==================================================  ON FILTER PROPERTY CHANGE ==================================================  */
const onFilterPropertyChange = (value: SelectableValue<string>) => {
  if (!value?.value) {return};
  
  onChange({ 
    ...query, 
    filterProperty: value.value 
  });
  runQueryIfChanged();
};

/* ==================================================  ON INCLUDE GROUP NAME ==================================================  */
const onIncludeGroupName = (event: ChangeEvent<HTMLInputElement>) => {
  onChange({ ...query, includeGroupName: event.currentTarget.checked })
  runQueryIfChanged()
}


/* ==================================================  ON INCLUDE DEVICE NAME ==================================================  */
const onIncludeDeviceName = (event: React.ChangeEvent<HTMLInputElement>) => {
  onChange({ ...query, includeDeviceName: event.currentTarget.checked })
  runQueryIfChanged()
}


/* ==================================================  ON INCLUDE SENSOR NAME ==================================================  */
const onIncludeSensorName = (event: ChangeEvent<HTMLInputElement>) => {
  onChange({ ...query, includeSensorName: event.currentTarget.checked })
  runQueryIfChanged()
}

/* ==================================================  ON MANUAL METHOD CHANGE ==================================================  */
const onManualMethodChange = (value: SelectableValue<string>) => {
  setManualMethod(value.value!);
  onChange({
    ...query,
    manualMethod: value.value,
  });
  runQueryIfChanged();
};


/* ==================================================  ON MANUAL OBJECT ID CHANGE ==================================================  */
const onManualObjectIdChange = (event: ChangeEvent<HTMLInputElement>) => {
  const value = event.currentTarget.value;
  setManualObjectId(value);
  onChange({
    ...query,
    manualObjectId: value,
  });
  runQueryIfChanged();
};



/* ================================================== DESTRUCTURING ================================================== */

// Set default streaming values with background updates enabled
useEffect(() => {
  if (query.isStreaming === undefined) {
    onChange({
      ...query,
      isStreaming: false,
      streamInterval: 2500, // Default interval 5ms (2,5 seconds)
      keepGraphsVisible: true, // Add this new option to keep charts visible during updates
    });
  }
}, [query, onChange]);

// Streaming section with backend integration and visibility options
const renderStreamingOptions = () => (
  <FieldSet label="Streaming Options">
    <Stack direction="row" gap={1}>
      <InlineField label="Enable Streaming" labelWidth={16}>
        <InlineSwitch 
          id='query-editor-is-stream'
          value={query.isStreaming || false}
          onChange={(e) => {
            const isStreaming = e.currentTarget.checked;
            const streamInterval = isStreaming ? (query.streamInterval || 2500) : undefined;
            onChange({ 
              ...query, 
              isStreaming,
              streamInterval,
              keepGraphsVisible: true, // Always keep graphs visible
            });
            // Run query to update backend state
            runQueryIfChanged();
          }} 
        />
      </InlineField>
      {query.isStreaming && (
        <>
          <InlineField label="Update Interval (ms)" labelWidth={20} tooltip="Refresh interval in milliseconds">
            <Input
              id='query-editor-stream-interval'
              type="number"
              value={query.streamInterval || 2500}
              onChange={(e) => {
                const interval = Math.max(0, Math.min(60000, parseInt(e.currentTarget.value, 10) || 2500));
                onChange({
                  ...query,
                  streamInterval: interval,
                });
                // Run query to update backend state
                runQueryIfChanged();
              }}
              placeholder="2500"
              min={0}
              max={60000}
            />
          </InlineField>
        </>
      )}
    </Stack>
  </FieldSet>
);

/* ================================================== RENDER ================================================== */
return (
  <Stack direction="column" gap={2}>
    <Stack direction="row" gap={2}>
      <Stack direction="column" gap={1}>
        <InlineField label="Query Type" labelWidth={20} grow>
          <Select
            id='query-editor-queryType'
            options={queryTypeOptions}
            value={query.queryType}
            onChange={onQueryTypeChange}
            width={47}
          />
        </InlineField>

        <InlineField label="Group" labelWidth={20} grow>
          <Select
            id='query-editor-group'
            isLoading={isLoading}
            options={groupOptions}
            value={selectedGroup}
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
            id='query-editor-device'
            isLoading={!lists.devices.length}
            options={deviceOptions}
            value={selectedDevice}
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
            id='query-editor-sensor'
            isLoading={!lists.sensors.length}
            options={sensorOptions}
            value={selectedSensor}
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
            id='query-editor-channel'
            key={sensorId}
            loadOptions={loadChannelOptions}
            defaultOptions={true}
            value={channelQuery.map(c => ({ 
              label: c, 
              value: c,
            }))}
            onChange={onChannelChange}
            width={47}
            placeholder={sensorId ? "Select Channel" : "First select a sensor"}
            isClearable
            isDisabled={!sensorId}
          />
        </InlineField>
      </Stack>
    </Stack>
   

    {/* Show display name options for both Metrics and Streaming */}
    {(isMetricsMode || query.isStreaming) && (
      <FieldSet label="Display Options">
        <Stack direction="row" gap={1}>
          <InlineField  label="Include Group" labelWidth={16}>
            <InlineSwitch id='query-editor-include-group' value={query.includeGroupName || false} onChange={onIncludeGroupName} />
          </InlineField>
          <InlineField label="Include Device" labelWidth={15}>
            <InlineSwitch id='query-editor-include-device' value={query.includeDeviceName || false} onChange={onIncludeDeviceName} />
          </InlineField>
          <InlineField label="Include Sensor" labelWidth={15}>
            <InlineSwitch  id='query-editor-include-device' value={query.includeSensorName || false} onChange={onIncludeSensorName} />
          </InlineField>
        </Stack>
      </FieldSet>
    )}

    {/* Options for Text and Raw modes */}
    {(isTextMode || isRawMode) && (
      <FieldSet label="Options">
        <Stack direction="row" gap={2}>
          <InlineField label="Property" labelWidth={16} tooltip="Select property type">
            <Select
              id='query-editor-property'
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
              id='query-editor-filterProperty'
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
              id='query-editor-manualMethod'
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
                runQueryIfChanged();
              }}
              isClearable
            />
          </InlineField>
          <InlineField label="Object ID" labelWidth={16} tooltip="Object ID from selected sensor">
            <Input
              id='query-editor-manualObjectId'
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

    {/* Always show streaming options */}
    {renderStreamingOptions()}

  </Stack>
)
}
