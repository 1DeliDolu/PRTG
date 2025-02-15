import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);

  // ✅ Check visibility of main fields in Query Editor
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Query Type' })).toBeVisible();
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Group' })).toBeVisible();
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Device' })).toBeVisible();
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Sensor' })).toBeVisible();
});

test('should trigger new query when group, device, and sensor fields are changed', async ({
  panelEditPage,
  readProvisionedDataSource,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);

  // ✅ select Query Type
  await panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Query Type' }).selectOption('Metrics');

  // ✅ select Group
  await panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Group' }).selectOption('Network Devices');

  // ✅ select Device
  await panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Device' }).selectOption('Router');

  // ✅ select Sensor
  await panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Sensor' }).selectOption('CPU Load');

  // ✅ Check if a new query is triggered
  const queryReq = panelEditPage.waitForQueryDataRequest();
  await expect(await queryReq).toBeTruthy();
});

test('data query should return expected values', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);

  // ✅ Select Query Type
  await panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Query Type' }).selectOption('Metrics');

  // ✅ Select Group, Device, Sensor
  await panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Group' }).selectOption('Network Devices');
  await panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Device' }).selectOption('Router');
  await panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Sensor' }).selectOption('CPU Load');

  // ✅ Select "Table" as visualization
  await panelEditPage.setVisualization('Table');

  // ✅ Paneli yenile ve beklenen değerlerin geldiğini doğrula
  await expect(panelEditPage.refreshPanel()).toBeOK();
  await expect(panelEditPage.panel.data).toContainText(['10', '20']);
});

// ✅ Test if error is returned when Query is missing
test('should show error message when sensor is missing', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);

  // ✅ Select Query Type
  await panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Query Type' }).selectOption('Metrics');

  // ✅ Select Group but don't select Sensor
  await panelEditPage.getQueryEditorRow('A').getByRole('combobox', { name: 'Group' }).selectOption('Network Devices');

  // ✅ Refresh panel and check for error message
  await expect(panelEditPage.refreshPanel()).not.toBeOK();
  await expect(panelEditPage).toHaveAlert('error', { hasText: 'Sensor is required' });
});
