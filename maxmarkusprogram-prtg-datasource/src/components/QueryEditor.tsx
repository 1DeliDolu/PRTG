import React, { useEffect, useState } from 'react'
import { InlineField, Select, Stack, FieldSet, InlineSwitch } from '@grafana/ui'
import { QueryEditorProps, SelectableValue } from '@grafana/data'
import { DataSource } from '../datasource'
import { MyDataSourceOptions, MyQuery, queryTypeOptions, QueryType, propertyList, filterPropertyList } from '../types'

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const isMetricsMode = query.queryType === QueryType.Metrics
  const isRawMode = query.queryType === QueryType.Raw
  const isTextMode = query.queryType === QueryType.Text

  const [group, setGroup] = useState<string>('')
  const [device, setDevice] = useState<string>('')
  //@ts-ignore
  const [sensor, setSensor] = useState<string>('')
  //@ts-ignore
  const [channel, setChannel] = useState<string>('')
  const [objid, setObjid] = useState<string>('')

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

  /* ############################################## FETCH GROUPS ####################################### */
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

  /* ########################################### FETCH DEVICES ####################################### */
  useEffect(() => {
    async function fetchDevices() {
      setIsLoading(true)
      try {
        const response = await datasource.getDevices()
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
        } else {
          console.error('Invalid response format:', response)
        }
      } catch (error) {
        console.error('Error fetching devices:', error)
      }
      setIsLoading(false)
    }
    fetchDevices()
  }, [datasource, group])

  /* ######################################## FETCH SENSOR ############################################### */
  useEffect(() => {
    async function fetchSensors() {
      setIsLoading(true)
      try {
        const response = await datasource.getSensors()
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
        } else {
          console.error('Invalid response format:', response)
        }
      } catch (error) {
        console.error('Error fetching sensors:', error)
      }
      setIsLoading(false)
    }
    fetchSensors()
  }, [datasource, device])

  /* ####################################### FETCH CHANNEL ############################################# */

  useEffect(() => {
    async function fetchChannels() {
      if (!objid) {
        return
      }

      setIsLoading(true)
      try {
        const response = await datasource.getChannels(objid)

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

    if (objid) {
      fetchChannels()
    }
  }, [datasource, objid])

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

  /* ######################################## QUERY  ############################################### */

  const onQueryTypeChange = (value: SelectableValue<QueryType>) => {
    onChange({
      ...query,
      queryType: value.value!,
    })
    onRunQuery()
  }

  const onGroupChange = (value: SelectableValue<string>) => {
    onChange({
      ...query,
      group: value.value!,
      device: '',
      sensor: '',
      channel: '',
      objid: '',
    })

    setGroup(value.value!)
    setDevice('')
    setSensor('')
    setChannel('')
    setObjid('')

    setLists((prev) => ({
      ...prev,
      devices: [],
      sensors: [],
      channels: [],
    }))
  }

  const onDeviceChange = (value: SelectableValue<string>) => {
    onChange({
      ...query,
      device: value.value!,
      sensor: '',
      channel: '',
    })
    setDevice(value.value!)
    setSensor('')
    setChannel('')

    setLists((prev) => ({
      ...prev,
      sensors: [],
      channels: [],
    }))
  }

  const findSensorObjid = async (sensorName: string) => {
    try {
      const response = await datasource.getSensors()
      if (response && Array.isArray(response.sensors)) {
        const sensor = response.sensors.find((s) => s.sensor === sensorName)
        if (sensor) {
          setObjid(sensor.objid.toString())
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

  const onSensorChange = async (value: SelectableValue<string>) => {
    const sensorObjid = await findSensorObjid(value.value!)
    onChange({
      ...query,
      sensor: value.value!,
      objid: sensorObjid,
      channel: '',
    })
    setSensor(value.value!)
    setObjid(sensorObjid)
    setChannel('')

    setLists((prev) => ({
      ...prev,
      channels: [],
    }))

    onRunQuery()
  }

  const onChannelChange = (value: SelectableValue<string>) => {
    onChange({ ...query, channel: value.value! })
    setChannel(value.value!)
    onRunQuery()
  }

  const onPropertyChange = (value: SelectableValue<string>) => {
    onChange({ ...query, property: value.value! })
    onRunQuery()
  }

  const onFilterPropertyChange = (value: SelectableValue<string>) => {
    onChange({ ...query, filterProperty: value.value! })
    onRunQuery()
  }

  const onIncludeGroupName = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, includeGroupName: e.currentTarget.checked })
    onRunQuery()
  }

  const onIncludeDeviceName = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, includeDeviceName: e.currentTarget.checked })
    onRunQuery()
  }

  const onIncludeSensorName = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, includeSensorName: e.currentTarget.checked })
    onRunQuery()
  }

  return (
    <Stack direction="column" gap={2}>
      {/* Erste Zeile: Query- und Sensorfelder */}
      <Stack direction="row" gap={2}>
        {/* Linke Spalte */}
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
              placeholder="Select Device or type '*'"
              isClearable
              isDisabled={!query.group}
            />
          </InlineField>
        </Stack>

        {/* Rechte Spalte */}
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

          {isMetricsMode && (
            <InlineField label="Channel" labelWidth={20} grow>
              <Select
                isLoading={!lists.channels.length}
                options={lists.channels}
                value={query.channel}
                onChange={onChannelChange}
                width={47}
                allowCustomValue
                placeholder="Select Channel or type '*'"
                isClearable
                isDisabled={!query.sensor}
              />
            </InlineField>
          )}
        </Stack>
      </Stack>

      {/* Zweite Zeile: Options */}
      <Stack direction="column" gap={1}>
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

        {isTextMode && (
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

        {isRawMode && (
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
      </Stack>
    </Stack>

  )
}
