import { test, expect } from '@playwright/test';
import path from 'path';

const SCREENSHOT_DIR = path.join(__dirname, '../../docs/screenshots');

// Helper to wait for animations to complete
async function waitForAnimations(page: any, ms = 500) {
  await page.waitForTimeout(ms);
}

test.describe('Marketing Screenshots', () => {
  test.beforeEach(async ({ page }) => {
    // Go to dashboard
    await page.goto('/');
    
    // Wait for WebSocket to connect and data to load
    await page.waitForSelector('.grid', { timeout: 15000 });
    await waitForAnimations(page, 1000);
  });

  test('hero - full dashboard', async ({ page }) => {
    // Set viewport to 1920x1080 for 4K-ready screenshot
    await page.setViewportSize({ width: 1920, height: 1080 });
    
    // Wait for all components to be visible
    await expect(page.locator('.grid').first()).toBeVisible();
    await expect(page.locator('table').or(page.getByText('No runs yet'))).toBeVisible();
    
    // Take full page screenshot
    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'hero-dashboard.png'),
      fullPage: true,
    });
  });

  test('stats cards component', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    
    // Locate stats cards section
    const statsGrid = page.locator('.grid').first();
    await expect(statsGrid).toBeVisible();
    
    // Take screenshot of just the stats cards area
    await statsGrid.screenshot({
      path: path.join(SCREENSHOT_DIR, 'stats-cards.png'),
    });
  });

  test('runs table component', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    
    // Find the runs table section (within the 2-column grid)
    const tableSection = page.locator('.lg\\:col-span-2').first();
    await expect(tableSection).toBeVisible();
    
    await tableSection.screenshot({
      path: path.join(SCREENSHOT_DIR, 'runs-table.png'),
    });
  });

  test('provider cards component', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    
    // Find provider health section
    const providerSection = page.locator('text=Provider Health').locator('..');
    await expect(providerSection).toBeVisible();
    
    await providerSection.screenshot({
      path: path.join(SCREENSHOT_DIR, 'provider-cards.png'),
    });
  });

  test('activity feed component', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    
    // Find activity feed section
    const activitySection = page.locator('text=Activity').locator('..');
    await expect(activitySection).toBeVisible();
    
    await activitySection.screenshot({
      path: path.join(SCREENSHOT_DIR, 'activity-feed.png'),
    });
  });

  test('mobile view', async ({ page }) => {
    // Set to iPhone 14 Pro size
    await page.setViewportSize({ width: 393, height: 852 });
    
    // Wait for mobile layout
    await waitForAnimations(page, 500);
    
    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'mobile-view.png'),
      fullPage: true,
    });
  });
});

test.describe('Cross-Browser Screenshots', () => {
  test('chrome desktop', async ({ page, browserName }) => {
    test.skip(browserName !== 'chromium', 'Chrome screenshot only');
    
    await page.goto('/');
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForSelector('.grid', { timeout: 15000 });
    await waitForAnimations(page, 1000);
    
    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'cross-browser/chrome.png'),
      fullPage: false,
    });
  });

  test('firefox desktop', async ({ page, browserName }) => {
    test.skip(browserName !== 'firefox', 'Firefox screenshot only');
    
    await page.goto('/');
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForSelector('.grid', { timeout: 15000 });
    await waitForAnimations(page, 1000);
    
    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'cross-browser/firefox.png'),
      fullPage: false,
    });
  });

  test('safari desktop', async ({ page, browserName }) => {
    test.skip(browserName !== 'webkit', 'Safari screenshot only');
    
    await page.goto('/');
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForSelector('.grid', { timeout: 15000 });
    await waitForAnimations(page, 1000);
    
    await page.screenshot({
      path: path.join(SCREENSHOT_DIR, 'cross-browser/safari.png'),
      fullPage: false,
    });
  });
});
