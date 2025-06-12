declare module '@grafana/plugin-e2e' {
    import { Page, APIRequestContext } from '@playwright/test';
  
    export interface GrafanaPage {
      ctx: any;
      pageArgs: any;
      navigate: (url: string) => Promise<void>;
      getByGrafanaSelector: (selector: string) => any;
    }
  
    export interface PanelEditPage {
      datasource: {
        set: (name: string) => Promise<void>;
      };
      setVisualization: (type: string) => Promise<void>;
    }
  
    export interface TestArgs {
      page: Page & GrafanaPage;
      request: APIRequestContext;
      panelEditPage: PanelEditPage;
      createDataSourceConfigPage: (options: { type: string }) => Promise<DataSourceConfigPage>;
      readProvisionedDataSource: (options: { fileName: string }) => Promise<any>;
      gotoDataSourceConfigPage: (options: { uid: string }) => Promise<DataSourceConfigPage>;
    }
  
    export interface DataSourceConfigPage extends GrafanaPage {
      saveAndTest(): Promise<any>;
    }
  
    export function test(name: string, fn: (args: TestArgs) => Promise<void>): void;
    export function test(name: string, options: any, fn: (args: TestArgs) => Promise<void>): void;
    
    export namespace test {
      function setTimeout(timeout: number): void;
      function skip(): void;
      function fixme(condition: boolean, reason: string): void;
    }
  
    export interface ExpectMatchers {
      toHaveAlert(type: 'error' | 'success' | 'warning', options?: { timeout?: number }): Promise<void>;
    }
  
    export { expect } from '@playwright/test';
  }
  