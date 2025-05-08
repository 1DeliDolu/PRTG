import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData , timezoneOptions } from '../types';
import { Select } from '@grafana/ui';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onPathChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        path: event.target.value,
      },
    });
  };

  // Secure field (only sent to the backend)
  const onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        apiKey: event.target.value,
      },
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
  // cachetime 
  const onCacheTimeChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        cacheTime: parseInt(event.target.value, 10),
      },
    });
  }

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
    <>
      <InlineField label="Path" labelWidth={14} interactive tooltip={'Json field returned to frontend'}>
        <Input
          id="config-editor-path"
          onChange={onPathChange}
          value={jsonData.path}
          placeholder="Enter the path, <your.prtg.server> without https://"
          width={60}
        />
      </InlineField>
      <InlineField label="API Key" labelWidth={14} interactive tooltip={'Secure json field (backend only)'}>
        <SecretInput
          required
          id="config-editor-api-key"
          isConfigured={secureJsonFields.apiKey}
          value={secureJsonData?.apiKey}
          placeholder="Enter your API key"
          width={60}
          onReset={onResetAPIKey}
          onChange={onAPIKeyChange}
        />
      </InlineField>
      <InlineField label="Cache Time" labelWidth={14} interactive tooltip={'Cache time in seconds'}>
        <Input
          id="config-editor-cache-time"
          onChange={onCacheTimeChange}
          value={jsonData.cacheTime}
          placeholder="Enter the cache time"
          width={60}
        />
      </InlineField>
      <InlineField label="Timezone" labelWidth={14} interactive tooltip={'Select the timezone'} required>
        <Select
          options={timezoneOptions}
          value={jsonData.timezone}
          onChange={onTimezoneChange}
          width={60}
        />
      </InlineField>
    </>
  );
}
