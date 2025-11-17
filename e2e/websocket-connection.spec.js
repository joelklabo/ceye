const { test, expect } = require('@playwright/test');

test.describe('WebSocket Connection', () => {
  test('should establish WebSocket connection and receive initial snapshot', async ({ page }) => {
    // Navigate to dashboard
    await page.goto('http://localhost:8080');
    
    // Wait for page to load
    await page.waitForLoadState('networkidle');
    
    // Check connection status shows "Connected"
    const connectionStatus = page.locator('#connectionStatus');
    await expect(connectionStatus).toContainText('Connected', { timeout: 10000 });
    
    // Check that we're NOT showing "Disconnected"
    await expect(connectionStatus).not.toContainText('Disconnected');
    
    // Check that last update timestamp is shown (not "No updates yet")
    const lastUpdate = page.locator('#lastUpdate');
    await expect(lastUpdate).not.toContainText('No updates yet', { timeout: 10000 });
    
    // Check that at least one stat is non-zero (we have demo data)
    const stats = await page.locator('.stat-value').allTextContents();
    const hasNonZeroStat = stats.some(stat => parseInt(stat) > 0);
    expect(hasNonZeroStat).toBeTruthy();
  });

  test('should receive periodic updates via WebSocket', async ({ page }) => {
    // Navigate to dashboard
    await page.goto('http://localhost:8080');
    
    // Wait for initial connection
    await page.waitForSelector('.connection-indicator.connected', { timeout: 10000 });
    
    // Get initial update timestamp
    const getLastUpdateText = () => page.locator('#lastUpdate').textContent();
    const initialUpdate = await getLastUpdateText();
    
    // Wait for an update (demo provider sends every 5 seconds)
    await page.waitForFunction(
      (initial) => {
        const current = document.querySelector('#lastUpdate').textContent;
        return current !== initial && !current.includes('No updates yet');
      },
      initialUpdate,
      { timeout: 10000 }
    );
    
    // Verify we got a new update
    const newUpdate = await getLastUpdateText();
    expect(newUpdate).not.toBe(initialUpdate);
    expect(newUpdate).not.toContain('No updates yet');
  });

  test('should show activity log with WebSocket messages', async ({ page }) => {
    // Navigate to dashboard
    await page.goto('http://localhost:8080');
    
    // Wait for connection
    await page.waitForSelector('.connection-indicator.connected', { timeout: 10000 });
    
    // Open activity log
    await page.click('#activityToggle');
    
    // Wait for activity log to be visible
    await page.waitForSelector('#activityLog[style*="display: block"]');
    
    // Check that activity log has entries (not just "Waiting for updates...")
    const activityItems = await page.locator('#activityLog .activity-item:not(.muted)').count();
    expect(activityItems).toBeGreaterThan(0);
  });
});
