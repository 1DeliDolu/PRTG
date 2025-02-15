import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

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

  return (
    <>
      <InlineField label="Path" labelWidth={14} interactive tooltip={'Json field returned to frontend'}>
        <Input
          id="config-editor-path"
          onChange={onPathChange}
          value={jsonData.path}
          placeholder="Enter the path, e.g. /api/v1"
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
    </>
  );
}
