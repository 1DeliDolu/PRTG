import React, { useEffect, useState } from 'react'
import { InlineField, Select, Stack, FieldSet, InlineSwitch, Input } from '@grafana/ui'
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
  //@ts-ignore
  const [sensor, setSensor] = useState<string>(query.sensor || '')
  //@ts-ignore
  const [channel, setChannel] = useState<string[]>(query.channels || [])
  const [sensorId, setSensorId] = useState<string>(query.sensorId || '')
  const [manualMethod, setManualMethod] = useState<string>(query.manualMethod || '');
  const [manualObjectId, setManualObjectId] = useState<string>(query.manualObjectId || '');


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
  //@ts-ignore
  const [groupId, setGroupId] = useState<string>('')
  //@ts-ignore
  const [deviceId, setDeviceId] = useState<string>('')


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
      if (!group) {return}; // Eğer group boşsa fetch yapmayı engelle
      
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
      if (!device) {return}; // Eğer device boşsa fetch yapmayı engelle

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
        return
      }

      setIsLoading(true)
      try {
        const response = await datasource.getChannels(sensorId)

        // Check if response is empty
        if (!response) {
          console.error('Empty response received')
          setLists((prev) => ({
            ...prev,
            channels: [],
          }))
          return
        }

        if (typeof response !== 'object') {
          console.error('Invalid response format:', response)
          return
        }

        if ('error' in response) {
          console.error('API Error:', response.error)
          return
        }

        if (!Array.isArray(response.values)) {
          console.error('Invalid channels format:', response)
          return
        }

        const channelOptions = Object.keys(response.values[0] || {})
          .filter((key) => key !== 'datetime')
          .map((key) => ({
            label: key,
            value: key,
          }))

        setLists((prev) => ({
          ...prev,
          channels: channelOptions,
        }))
      } catch (error) {
        console.error('Error fetching channels:', error)
        setLists((prev) => ({
          ...prev,
          channels: [],
        }))
      }
      setIsLoading(false)
    }

    if (sensorId) {
      fetchChannels()
    }
  }, [datasource, sensorId])

  useEffect(() => {
    if (isTextMode || isRawMode) {
      const propertyOptions: Array<SelectableValue<string>> = propertyList.map((item) => ({
        label: item.visible_name,
        value: item.name,
      }))

      const filterPropertyOptions: Array<SelectableValue<string>> = filterPropertyList.map((item) => ({
        label: item.visible_name,
        value: item.name,
      }))

      setLists((prev) => ({
        ...prev,
        properties: propertyOptions,
        filterProperties: filterPropertyOptions,
      }))
    }
  }, [isTextMode, isRawMode])

  /* ==================================================  INITIAL VALUES  ================================================== */
  useEffect(() => {
    setGroup(query.group || '')
    setDevice(query.device || '')
    setSensor(query.sensor || '')
    setChannel(query.channels || [])
    setSensorId(query.sensorId || '')
  }, [query.group, query.device, query.sensor, query.channels, query.sensorId]) 


  /* ==================================================  QUERY  ==================================================  */

  /* ==================================================  ONQUERYTYPESCHANGE ==================================================  */
    const onQueryTypeChange = (value: SelectableValue<QueryType>) => {
    onChange({
      ...query,
      queryType: value.value!,
    })
    onRunQuery()
  }

  /* ==================================================  FIND GROUP ID ================================================= */
  const findGroupId = async (groupName: string) => {
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
  }

  const findDeviceId = async (deviceName: string) => {
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
  }

  const findSensorObjid = async (sensorName: string) => {
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
  }

  /* ==================================================  ONGROUPCHANGE ==================================================  */
  const onGroupChange = async (value: SelectableValue<string>) => {
    const groupObjId = await findGroupId(value.value!)

    onChange({
      ...query,
      group: value.value!,
      groupId: groupObjId,
      device: '',
      deviceId: '',
      sensor: '',
      sensorId: '',
      channel: '',
    })

    setGroup(value.value!)
    setGroupId(groupObjId)
    setDevice('')
    setDeviceId('')
    setSensor('')
    setSensorId('')
    setChannel([])

    setLists((prev) => ({
      ...prev,
      devices: [],
      sensors: [],
      channels: [],
    }))
    onRunQuery()
  }

  const onDeviceChange = async (value: SelectableValue<string>) => {
    const deviceObjId = await findDeviceId(value.value!)

    onChange({
      ...query,
      device: value.value!,
      deviceId: deviceObjId,
      sensor: '',
      sensorId: '',
      channel: '',
    })

    setDevice(value.value!)
    setDeviceId(deviceObjId)
    setSensor('')
    setSensorId('')
    setChannel([])

    setLists((prev) => ({
      ...prev,
      sensors: [],
      channels: [],
    }))
    onRunQuery()
  }


/* ==================================================  ONSENSORCHANGE ==================================================  */
  const onSensorChange = async (value: SelectableValue<string>) => {
    const sensorObjId = await findSensorObjid(value.value!)

    if (isMetricsMode) {
      onChange({
        ...query,
        sensor: value.value!,
        sensorId: sensorObjId,
        channel: '',
      })
      setChannel([])
    } else if (isManualMode) {

      onChange({
        ...query,
        sensor: value.value!,
        sensorId: sensorObjId,
        manualObjectId: sensorObjId,
      })
      setManualObjectId(sensorObjId)
    } else {
      onChange({
        ...query,
        sensor: value.value!,
        sensorId: sensorObjId,
      })
    }

    setSensor(value.value!)
    setSensorId(sensorObjId)

    if (isMetricsMode) {
      setLists((prev) => ({
        ...prev,
        channels: [],
      }))
    }
    onRunQuery()
  }

  /* ==================================================  ONCHANNELCHANGE ==================================================  */
  const onChannelChange = (value: SelectableValue<string> | Array<SelectableValue<string>>) => {
    const selectedChannels = Array.isArray(value) ? value.map(v => v.value || '') : [];
    onChange({
      ...query,
      channels: selectedChannels,
      channel: selectedChannels[0] || '',
    });

    setChannel(selectedChannels);
    onRunQuery();
  };

/* ==================================================  ONPROPERTYCHANGE ==================================================  */
  const onPropertyChange = (value: SelectableValue<string>) => {
    onChange({ ...query, property: value.value! })
    onRunQuery()
  }


  /* ==================================================  ONFILTERPROPERTYCHANGE ================================================= */
  const onFilterPropertyChange = (value: SelectableValue<string>) => {
    onChange({ ...query, filterProperty: value.value! })
    onRunQuery()
  }


  /* ==================================================  ONINCLUDEGROUPNAME ==================================================  */
  const onIncludeGroupName = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, includeGroupName: e.currentTarget.checked })
    onRunQuery()
  }


  /* ==================================================  ONINCLUDEDEVICENAME ==================================================  */
  const onIncludeDeviceName = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, includeDeviceName: e.currentTarget.checked })
    onRunQuery()
  }


  /* ==================================================  ONINCLUDESENSORNAME ==================================================  */
  const onIncludeSensorName = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, includeSensorName: e.currentTarget.checked })
    onRunQuery()
  }

  /* ==================================================  ONMANUALMETHODCHANGE ==================================================  */
  const onManualMethodChange = (value: SelectableValue<string>) => {
    setManualMethod(value.value!);
    onChange({
      ...query,
      manualMethod: value.value,
    });
    onRunQuery();
  };


  /* ==================================================  ONMANUALOBJECTIDCHANGE ==================================================  */
  const onManualObjectIdChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.currentTarget.value;
    setManualObjectId(value);
    onChange({
      ...query,
      manualObjectId: value,
    });
    onRunQuery();
  };

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
              options={lists.groups}
              value={query.group}
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
              options={lists.devices}
              value={query.device}
              onChange={onDeviceChange}
              width={47}
              allowCustomValue
              isClearable
              isDisabled={!query.group}
              placeholder="Select Device or type '*'"
            />
          </InlineField>
        </Stack>

        <Stack direction="column" gap={2}>
          <InlineField label="Sensor" labelWidth={20} grow>
            <Select
              isLoading={!lists.sensors.length}
              options={lists.sensors}
              value={query.sensor}
              onChange={onSensorChange}
              width={47}
              allowCustomValue
              placeholder="Select Sensor or type '*'"
              isClearable
              isDisabled={!query.device}
            />
          </InlineField>
          <InlineField label="Channel" labelWidth={20} grow>
            <Select
              isLoading={!lists.channels.length}
              options={lists.channels}
              value={(query.channels || []).map(c => ({ label: c, value: c })) || []}
              onChange={onChannelChange}
              width={47}
              allowCustomValue
              placeholder="Select Channel"
              isClearable
              isMulti={true}
              isDisabled={!query.sensor}
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
          </Stack>
        </FieldSet>
      )}

      {/* Options for Text and Raw modes */}
      {(isTextMode || isRawMode) && (
        <FieldSet label="Options">
          <Stack direction="row" gap={1}>
            <InlineField label="Property" labelWidth={16}>
              <Select
                options={lists.properties}
                value={query.property}
                onChange={onPropertyChange}
                width={32}
              />
            </InlineField>
            <InlineField label="Filter Property" labelWidth={16}>
              <Select
                options={lists.filterProperties}
                value={query.filterProperty}
                onChange={onFilterPropertyChange}
                width={32}
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
                allowCustomValue
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

