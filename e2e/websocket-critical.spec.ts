import { test, expect } from '@playwright/test';

test.describe('WebSocket Connection - Critical Issues', () => {
  test.beforeEach(async ({ page }) => {
    // Start with a fresh page
    await page.goto('http://localhost:8080');
  });

  test('should NOT show OFFLINE status constantly', async ({ page }) => {
    // Wait for React app to render
    await page.waitForSelector('[data-testid="connection-indicator"]', { timeout: 5000 });
    
    // Wait a moment for WebSocket to attempt connection
    await page.waitForTimeout(2000);
    
    // Check connection status - should be LIVE, not OFFLINE
    const indicator = page.locator('[data-testid="connection-indicator"]');
    const text = await indicator.textContent();
    
    // This test will FAIL if WebSocket is broken
    expect(text).toContain('LIVE');
    expect(text).not.toContain('OFFLINE');
  });

  test('should NOT show "Max reconnection attempts" error', async ({ page }) => {
    // Listen for console errors
    const consoleErrors: string[] = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });
    
    // Wait for page to load and try to connect
    await page.waitForTimeout(5000);
    
    // Check for the specific error
    const hasReconnectionError = consoleErrors.some(err => 
      err.includes('Max reconnection attempts') || 
      err.includes('reconnection')
    );
    
    expect(hasReconnectionError).toBe(false);
  });

  test('should establish WebSocket connection within 3 seconds', async ({ page }) => {
    const startTime = Date.now();
    
    // Wait for LIVE status
    await page.waitForSelector('[data-testid="connection-indicator"]:has-text("LIVE")', { 
      timeout: 3000 
    });
    
    const connectionTime = Date.now() - startTime;
    
    // Should connect quickly
    expect(connectionTime).toBeLessThan(3000);
  });

  test('should receive WebSocket messages', async ({ page }) => {
    // Wait for connection
    await page.waitForSelector('[data-testid="connection-indicator"]:has-text("LIVE")', { 
      timeout: 5000 
    });
    
    // Check for activity feed updates
    const activityFeed = page.locator('[data-testid="activity-feed"]');
    await expect(activityFeed).toBeVisible({ timeout: 5000 });
    
    // Wait for at least one activity item
    const activityItems = page.locator('[data-testid="activity-item"]');
    await expect(activityItems.first()).toBeVisible({ timeout: 10000 });
    
    // Should have received messages
    const count = await activityItems.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should update stats from WebSocket data', async ({ page }) => {
    // Wait for connection
    await page.waitForSelector('[data-testid="connection-indicator"]:has-text("LIVE")', { 
      timeout: 5000 
    });
    
    // Get stats cards
    const totalRuns = page.locator('[data-testid="stat-total-runs"]');
    const activeRuns = page.locator('[data-testid="stat-active-runs"]');
    
    // Should show stats (not 0 if demo is running)
    await expect(totalRuns).toBeVisible({ timeout: 3000 });
    await expect(activeRuns).toBeVisible({ timeout: 3000 });
    
    // Get values
    const totalText = await totalRuns.textContent();
    const activeText = await activeRuns.textContent();
    
    // Should have numbers
    expect(totalText).toMatch(/\d+/);
    expect(activeText).toMatch(/\d+/);
  });
});
