import { test, expect } from '@grafana/plugin-e2e';

test.describe('PRTG Datasource Plugin', () => {
  test('should display datasource configuration page', async ({ page, selectors }) => {
    await page.goto('/datasources/new');
    
    // Look for PRTG datasource option using the correct selector
    const datasourceOption = page.locator(selectors.pages.AddDataSource.dataSourcePluginsV2('PRTG'));
    await expect(datasourceOption).toBeVisible();
  });

  test('should configure PRTG datasource', async ({ page, selectors }) => {
    // Navigate to datasource configuration
    await page.goto('/datasources/new');
    
    // Select PRTG datasource
    await page.locator(selectors.pages.AddDataSource.dataSourcePluginsV2('PRTG')).click();
    
    // Fill in configuration
    await page.fill('[data-testid="prtg-url-input"]', 'https://your-prtg-server.com');
    await page.fill('[data-testid="prtg-username-input"]', 'testuser');
    await page.fill('[data-testid="prtg-password-input"]', 'testpassword');
    
    // Save configuration
    await page.click(selectors.pages.DataSource.saveAndTest);
    
    // Verify success message
    await expect(page.locator('[data-testid="data-testid Alert success"]')).toBeVisible();
  });

  test('should create a query', async ({ page }) => {
    // Navigate to explore page
    await page.goto('/explore');
    
    // Select PRTG datasource using direct selector
    await page.click('[data-testid="data-source-picker"]');
    await page.click('text=PRTG');
    
    // Configure query
    await page.fill('[data-testid="sensor-id-input"]', '12345');
    
    // Run query using direct selector
    await page.click('[data-testid="run-query"]');
    
    // Verify results
    await expect(page.locator('[data-testid="graph-container"]')).toBeVisible();
  });
});