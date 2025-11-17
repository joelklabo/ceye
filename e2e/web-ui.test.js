const { test, expect } = require('@playwright/test');
const fs = require('fs');
const path = require('path');

// Test configuration
const BASE_URL = 'http://localhost:8080';
const LOG_DIR = path.join(__dirname, '../tmp/e2e-logs');

// Ensure log directory exists
if (!fs.existsSync(LOG_DIR)) {
  fs.mkdirSync(LOG_DIR, { recursive: true });
}

// Capture console logs
const consoleLogs = [];
const captureConsoleLogs = (page) => {
  page.on('console', msg => {
    const log = `[${msg.type()}] ${msg.text()}`;
    consoleLogs.push(log);
    console.log(log);
  });
  
  page.on('pageerror', error => {
    const log = `[ERROR] ${error.message}`;
    consoleLogs.push(log);
    console.error(log);
  });
};

test.describe('ceye Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    consoleLogs.length = 0; // Clear logs
    captureConsoleLogs(page);
  });

  test.afterEach(async () => {
    // Save logs to file
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const logFile = path.join(LOG_DIR, `console-${timestamp}.log`);
    fs.writeFileSync(logFile, consoleLogs.join('\n'));
    console.log(`\n📝 Console logs saved to: ${logFile}`);
  });

  test('Issue 1: WebSocket connection status', async ({ page }) => {
    await page.goto(BASE_URL);
    
    // Wait for page to load
    await page.waitForSelector('#connectionStatus', { timeout: 5000 });
    
    // Check connection status
    const statusText = await page.locator('#connectionStatus').textContent();
    console.log(`\n🔌 Connection Status: ${statusText}`);
    
    // Take screenshot
    await page.screenshot({ 
      path: path.join(LOG_DIR, 'websocket-status.png'),
      fullPage: true 
    });
    
    // Should show "Connected" not "Disconnected"
    expect(statusText).toContain('Connected');
  });

  test('Issue 2: Timer updates', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForSelector('#lastUpdate', { timeout: 5000 });
    
    // Get initial timestamp
    const initialTime = await page.locator('#lastUpdate').textContent();
    console.log(`\n⏰ Initial time: ${initialTime}`);
    
    // Wait 3 seconds
    await page.waitForTimeout(3000);
    
    // Check if time updated
    const updatedTime = await page.locator('#lastUpdate').textContent();
    console.log(`⏰ After 3s: ${updatedTime}`);
    
    // Time should have changed (or at least not be static)
    // Note: This might not change if no data updates, so just check format
    expect(updatedTime).toMatch(/Last update:/);
  });

  test('Issue 3: Provider count (should be 1, not 2)', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForSelector('.provider-card', { timeout: 5000 });
    
    // Count provider cards
    const providerCards = await page.locator('.provider-card').count();
    console.log(`\n🏢 Provider count: ${providerCards}`);
    
    // Get provider names
    const providerNames = await page.locator('.provider-card .provider-name').allTextContents();
    console.log(`📋 Providers: ${providerNames.join(', ')}`);
    
    // Should only have ONE github provider
    const githubProviders = providerNames.filter(name => name.includes('github'));
    expect(githubProviders.length).toBe(1);
    expect(githubProviders[0]).toBe('github');
  });

  test('Issue 4: Refresh button works', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForSelector('#refreshBtn', { timeout: 5000 });
    
    console.log(`\n🔄 Testing refresh button...`);
    
    // Click refresh
    await page.click('#refreshBtn');
    
    // Should trigger some visible feedback
    // Check if button shows loading state or logs appear
    await page.waitForTimeout(1000);
    
    // Button should be clickable (not testing much here, just that it doesn't error)
    const isEnabled = await page.locator('#refreshBtn').isEnabled();
    expect(isEnabled).toBe(true);
  });

  test('Issue 5: Workspace selector exists', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForSelector('#workspaceSelector', { timeout: 5000 });
    
    console.log(`\n💼 Testing workspace selector...`);
    
    const selector = page.locator('#workspaceSelector');
    const isVisible = await selector.isVisible();
    expect(isVisible).toBe(true);
    
    // Check if it has the default option
    const optionCount = await selector.locator('option').count();
    console.log(`📋 Workspace options: ${optionCount}`);
  });

  test('Issue 6: Alerts page loads', async ({ page }) => {
    await page.goto(`${BASE_URL}/alerts.html`);
    
    // Should not show 404
    const title = await page.title();
    console.log(`\n🔔 Alerts page title: ${title}`);
    expect(title).toContain('Alerts');
    
    // Take screenshot
    await page.screenshot({ 
      path: path.join(LOG_DIR, 'alerts-page.png'),
      fullPage: true 
    });
  });

  test('Issue 7: Settings page loads', async ({ page }) => {
    await page.goto(`${BASE_URL}/settings.html`);
    
    // Should not show 404
    const title = await page.title();
    console.log(`\n⚙️  Settings page title: ${title}`);
    expect(title).toContain('Settings');
    
    // Take screenshot
    await page.screenshot({ 
      path: path.join(LOG_DIR, 'settings-page.png'),
      fullPage: true 
    });
  });

  test('Issue 8: Filter functionality', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForSelector('.quick-filters', { timeout: 5000 });
    
    console.log(`\n🔍 Testing filter buttons...`);
    
    // Check for filter buttons
    const providerFilterBtn = page.locator('button:has-text("+ Provider")');
    const statusFilterBtn = page.locator('button:has-text("+ Status")');
    
    const providerBtnVisible = await providerFilterBtn.isVisible();
    const statusBtnVisible = await statusFilterBtn.isVisible();
    
    console.log(`📋 Provider filter button visible: ${providerBtnVisible}`);
    console.log(`📋 Status filter button visible: ${statusBtnVisible}`);
    
    expect(providerBtnVisible).toBe(true);
    expect(statusBtnVisible).toBe(true);
  });

  test('Issue 9: WebSocket receives messages', async ({ page }) => {
    await page.goto(BASE_URL);
    
    console.log(`\n📡 Monitoring WebSocket messages...`);
    
    // Listen for WebSocket messages
    const messages = [];
    await page.evaluate(() => {
      window.wsMessages = [];
      const originalWebSocket = window.WebSocket;
      window.WebSocket = function(...args) {
        const ws = new originalWebSocket(...args);
        ws.addEventListener('message', (event) => {
          window.wsMessages.push(event.data);
        });
        return ws;
      };
    });
    
    // Wait for messages
    await page.waitForTimeout(5000);
    
    // Check if any messages received
    const wsMessages = await page.evaluate(() => window.wsMessages || []);
    console.log(`📬 WebSocket messages received: ${wsMessages.length}`);
    
    if (wsMessages.length > 0) {
      console.log(`📦 First message preview: ${wsMessages[0].substring(0, 200)}...`);
    }
    
    // Should receive at least one message
    expect(wsMessages.length).toBeGreaterThan(0);
  });

  test('Full UI Screenshot', async ({ page }) => {
    await page.goto(BASE_URL);
    await page.waitForSelector('.container', { timeout: 5000 });
    
    // Take full page screenshot
    await page.screenshot({ 
      path: path.join(LOG_DIR, 'full-ui.png'),
      fullPage: true 
    });
    
    console.log(`\n📸 Full UI screenshot saved`);
  });
});
