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
  const [channel, setChannel] = useState<string[]>([])
  const [sensorId, setSensorId] = useState<string>('')

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

  /* ######################################## QUERY  ############################################### */

  const onQueryTypeChange = (value: SelectableValue<QueryType>) => {
    // Mevcut query'nin diğer değerlerini koruyarak sadece tipini değiştir
    onChange({
      ...query,
      queryType: value.value!,
    })
    onRunQuery()
  }

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
              placeholder="Select Device or type '*'"
              isClearable
              isDisabled={!query.group}
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

          {/* Channel seçimini sadece metrics modunda göster */}
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
      
      {/* Metrics modu için options */}
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

      {/* Text ve Raw modları için options */}
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
      {/* query selbt burada ben kendim urls getsensordteail vb veriler girip bir tabe panel olusttumak istiyorum . /api/getobjectproperty.htm?id=objectid&name=propertyname&show=text , /api/getsensordetails.xml?id=sensorid ,/api/getstatus.htm?id=0   */}

    </Stack>
  )
}

