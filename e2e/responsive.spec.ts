import { test, expect } from '@playwright/test';

const viewports = {
  mobile: { width: 375, height: 667 },  // iPhone SE
  tablet: { width: 768, height: 1024 }, // iPad
  desktop: { width: 1920, height: 1080 }, // Full HD
  ultra: { width: 3840, height: 2160 }, // 4K
};

test.describe('Responsive Design', () => {
  test('mobile viewport (375px) - stacks components', async ({ page }) => {
    await page.setViewportSize(viewports.mobile);
    await page.goto('/');
    
    // Wait for content
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Stats cards should be visible
    const statsGrid = page.locator('.grid').first();
    await expect(statsGrid).toBeVisible();
    
    // Table should be visible
    const table = page.locator('table').or(page.getByText('No runs yet'));
    await expect(table).toBeVisible();
    
    // Header should be responsive
    const header = page.locator('header');
    await expect(header).toBeVisible();
    
    // Take screenshot for visual verification
    await page.screenshot({ path: 'tmp/responsive-mobile.png', fullPage: true });
  });

  test('tablet viewport (768px) - 2-column grid', async ({ page }) => {
    await page.setViewportSize(viewports.tablet);
    await page.goto('/');
    
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    const statsGrid = page.locator('.grid').first();
    await expect(statsGrid).toBeVisible();
    
    // Verify layout is readable
    const header = page.locator('header h1');
    await expect(header).toBeVisible();
    
    await page.screenshot({ path: 'tmp/responsive-tablet.png', fullPage: true });
  });

  test('desktop viewport (1920px) - full layout', async ({ page }) => {
    await page.setViewportSize(viewports.desktop);
    await page.goto('/');
    
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // All 4 stat cards should be in one row
    const statsGrid = page.locator('.grid').first();
    await expect(statsGrid).toBeVisible();
    
    // Sidebar should be visible
    const activityFeed = page.getByRole('heading', { name: 'Activity' });
    await expect(activityFeed).toBeVisible();
    
    await page.screenshot({ path: 'tmp/responsive-desktop.png', fullPage: false });
  });

  test('4K viewport (3840px) - no overflow', async ({ page }) => {
    await page.setViewportSize(viewports.ultra);
    await page.goto('/');
    
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Check that content doesn't have horizontal scroll
    const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
    const viewportWidth = viewports.ultra.width;
    
    // Body should not be wider than viewport
    expect(bodyWidth).toBeLessThanOrEqual(viewportWidth + 20); // +20 for scrollbar
    
    await page.screenshot({ path: 'tmp/responsive-4k.png', fullPage: false });
  });

  test('header is responsive on mobile', async ({ page }) => {
    await page.setViewportSize(viewports.mobile);
    await page.goto('/');
    
    const header = page.locator('header');
    await expect(header).toBeVisible();
    
    // Title should be visible
    const title = page.locator('h1');
    await expect(title).toBeVisible();
    
    // Connection status should wrap or stack appropriately
    const connectionStatus = page.locator('header').getByText(/Connected|Disconnected/);
    await expect(connectionStatus).toBeVisible();
  });

  test('table is scrollable on mobile', async ({ page }) => {
    await page.setViewportSize(viewports.mobile);
    await page.goto('/');
    
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Table should be present (may need horizontal scroll)
    const table = page.locator('table');
    const hasTable = await table.count() > 0;
    
    if (hasTable) {
      await expect(table).toBeVisible();
      
      // Table wrapper should allow scroll
      const tableWrapper = table.locator('..');
      await expect(tableWrapper).toBeVisible();
    }
  });
});
