import { test as setup } from '@playwright/test';

const authFile = 'playwright/.auth/admin.json';

setup('authenticate', async ({ page }) => {
  // Create the auth directory if it doesn't exist using Playwright's built-in capabilities
  const path = require('path');
  const fs = require('fs');

  const authDir = path.dirname(authFile);
  if (!fs.existsSync(authDir)) {
    fs.mkdirSync(authDir, { recursive: true });
  }

  try {
    console.log('🔐 Starting authentication process...');

    // First check if Grafana is accessible
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const currentUrl = page.url();
    console.log(`📍 Current URL: ${currentUrl}`);

    // Check if we're already authenticated (anonymous access enabled)
    const isOnLoginPage =
      currentUrl.includes('/login') ||
      (await page.locator('input[type="password"]').isVisible());

    if (!isOnLoginPage) {
      console.log('✅ Already authenticated or anonymous access enabled');
      await page.context().storageState({ path: authFile });
      return;
    }

    console.log('🔑 Login required, attempting authentication...');

    // Navigate to login page if not already there
    if (!currentUrl.includes('/login')) {
      await page.goto('/login');
      await page.waitForLoadState('networkidle');
    }

    // Try multiple selectors for username field
    const usernameSelectors = [
      'input[name="user"]',
      'input[placeholder*="email"]',
      'input[placeholder*="username"]',
      'input[aria-label*="username"]',
      'input[aria-label*="email"]',
      '[data-testid="username"]',
      '[data-testid="email"]',
    ];

    const passwordSelectors = [
      'input[name="password"]',
      'input[type="password"]',
      'input[aria-label*="password"]',
      '[data-testid="password"]',
    ];

    let usernameField, passwordField;

    // Find username field
    for (const selector of usernameSelectors) {
      try {
        const field = page.locator(selector).first();
        if (await field.isVisible({ timeout: 2000 })) {
          usernameField = field;
          console.log(`📝 Found username field: ${selector}`);
          break;
        }
      } catch (e) {
        // Continue to next selector
      }
    }

    // Find password field
    for (const selector of passwordSelectors) {
      try {
        const field = page.locator(selector).first();
        if (await field.isVisible({ timeout: 2000 })) {
          passwordField = field;
          console.log(`🔒 Found password field: ${selector}`);
          break;
        }
      } catch (e) {
        // Continue to next selector
      }
    }

    if (!usernameField || !passwordField) {
      throw new Error('❌ Could not find login form fields');
    }

    // Fill login form
    await usernameField.fill('admin');
    await passwordField.fill('admin');

    console.log('✏️ Filled login credentials');

    // Find and click submit button
    const submitSelectors = [
      'button[type="submit"]',
      'input[type="submit"]',
      'button:has-text("Log in")',
      'button:has-text("Sign in")',
      'button:has-text("Login")',
      '[data-testid="login-button"]',
    ];

    let submitButton;
    for (const selector of submitSelectors) {
      try {
        const button = page.locator(selector).first();
        if (await button.isVisible({ timeout: 2000 })) {
          submitButton = button;
          console.log(`🎯 Found submit button: ${selector}`);
          break;
        }
      } catch (e) {
        // Continue to next selector
      }
    }

    if (!submitButton) {
      // Try pressing Enter as fallback
      console.log('⏎ Submit button not found, trying Enter key');
      await passwordField.press('Enter');
    } else {
      await submitButton.click();
    }

    console.log('🚀 Submitted login form');

    // Wait for successful login with multiple indicators
    const successIndicators = [
      'div.sidemenu',
      '[data-testid="sidemenu"]',
      'nav[aria-label*="menu"]',
      '.page-container',
      '.main-view',
      '[data-testid="data-testid-sidemenu"]',
      '.grafana-app',
      '.main-view .page-container',
    ];

    let loginSuccessful = false;

    // Try different success indicators with longer timeout
    for (const indicator of successIndicators) {
      try {
        console.log(`🔍 Looking for success indicator: ${indicator}`);
        await page.waitForSelector(indicator, { timeout: 15000 });
        loginSuccessful = true;
        console.log(`✅ Found success indicator: ${indicator}`);
        break;
      } catch (e) {
        console.log(`❌ Indicator not found: ${indicator}`);
        // Continue to next indicator
      }
    }

    // Alternative: check URL change
    if (!loginSuccessful) {
      console.log('🔍 Checking URL for successful redirect...');
      try {
        await page.waitForFunction(
          () => !window.location.href.includes('/login'),
          { timeout: 15000 }
        );
        loginSuccessful = true;
        console.log('✅ URL indicates successful login');
      } catch (e) {
        console.log('❌ URL check failed');
      }
    }

    // Final check
    const finalUrl = page.url();
    console.log(`📍 Final URL: ${finalUrl}`);

    if (finalUrl.includes('/login') || finalUrl.includes('/auth')) {
      throw new Error('❌ Authentication failed - still on login/auth page');
    }

    // Save authentication state
    await page.context().storageState({ path: authFile });
    console.log('💾 Saved authentication state');
  } catch (error) {
    console.error(
      '❌ Authentication error:',
      error instanceof Error ? error.message : String(error)
    );

    // Create screenshots directory if it doesn't exist
    const screenshotDir = 'tests/screenshots';
    const fs = require('fs');
    if (!fs.existsSync(screenshotDir)) {
      fs.mkdirSync(screenshotDir, { recursive: true });
    }

    // Take screenshot for debugging
    const timestamp = Date.now();
    await page.screenshot({
      path: `${screenshotDir}/auth-error-${timestamp}.png`,
      fullPage: true,
    });
    console.log(`📸 Screenshot saved: auth-error-${timestamp}.png`);

    // Check if we can proceed anyway (maybe anonymous access works)
    const currentUrl = page.url();

    if (!currentUrl.includes('/login') && !currentUrl.includes('/auth')) {
      console.log('⚠️ Error occurred but not on login page, proceeding...');

      // Create a basic auth state
      const basicAuthState = {
        cookies: [],
        origins: [
          {
            origin: 'http://localhost:3000',
            localStorage: [],
          },
        ],
      };

      fs.writeFileSync(authFile, JSON.stringify(basicAuthState, null, 2));
      console.log('💾 Created fallback auth state');
      return;
    }

    throw error;
  }
});
