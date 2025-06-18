/** @jsx React.createElement */
import React, { ChangeEvent } from 'react';
import { Combobox, InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';
import { timezoneOptions } from '../timezone'

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> { }

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onPathChange = (event: ChangeEvent<HTMLInputElement>) => {
    const updatedJsonData = { ...options.jsonData };
    updatedJsonData.path = event.target.value;

    onOptionsChange({
      ...options,
      jsonData: updatedJsonData,
    });
  };

  const onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const updatedSecureJsonData = { ...options.secureJsonData };
    updatedSecureJsonData.apiKey = event.target.value;

    onOptionsChange({
      ...options,
      secureJsonData: updatedSecureJsonData,
    });
  };

  const onResetAPIKey = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: '',
      },
    });
  };

  const onCacheTimeChange = (event: ChangeEvent<HTMLInputElement>) => {
    const updatedJsonData = { ...options.jsonData };
    const inputValue = event.target.value;

    if (inputValue === '') {
      updatedJsonData.cacheTime = 60; // Default to 60 seconds
    } else {
      const value = parseInt(inputValue, 10);
      // Set cache time, ensuring minimum value of 10 seconds
      if (!isNaN(value) && value >= 10) {
        updatedJsonData.cacheTime = value;
      } else {
        // Don't update for invalid values
        return;
      }
    }

    onOptionsChange({
      ...options,
      jsonData: updatedJsonData,
    });
  };
  // Timezone selection handler
  const onTimezoneChange = (selectedOption: any) => {
    // Update the timezone directly when selection changes
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        timezone: selectedOption.value,
      },
    });
  };

  return (
    <div>
      <InlineField label="Path" labelWidth={14} interactive tooltip="Json field returned to frontend">
        <Input
          id="config-editor-path"
          onChange={onPathChange}
          value={jsonData.path || ''}
          placeholder="Enter the path, <your.prtg.server> without https://"
          width={60}
        />
      </InlineField>
      <InlineField label="API Key" labelWidth={14} interactive tooltip="Secure json field (backend only)">
        <SecretInput
          required
          id="config-editor-api-key"
          isConfigured={secureJsonFields?.apiKey || false}
          value={secureJsonData?.apiKey || ''}
          placeholder="Enter your API key"
          width={60}
          onReset={onResetAPIKey}
          onChange={onAPIKeyChange}
        />
      </InlineField>
      <InlineField label="Cache Time" labelWidth={14} interactive tooltip="Cache time in seconds">
        <Input
          id="config-editor-cache-time"
          onChange={onCacheTimeChange}
          value={jsonData.cacheTime || 60}
          placeholder="Enter the cache time in seconds"
          width={60}
          type="number"
          min={10}
        />
      </InlineField>
      <InlineField label="Timezone" labelWidth={14} interactive tooltip={'Select the timezone'} required>
        <Combobox
          options={timezoneOptions}
          value={jsonData.timezone}
          onChange={onTimezoneChange}
          width={60}
        />
      </InlineField>
    </div>
  );
}