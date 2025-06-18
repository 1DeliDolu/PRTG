// Type definitions for @grafana/plugin-e2e
declare module '@grafana/plugin-e2e' {
  import { Page } from '@playwright/test';
  
  export interface TestContext {
    page: Page;
    selectors: any;
  }
  
  export const test: {
    describe: (name: string, fn: () => void) => void;
    (name: string, fn: (context: TestContext) => Promise<void>): void;
  };
  
  export const expect: any;
}