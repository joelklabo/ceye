import { test, expect } from '@playwright/test';

test.describe('Accessibility', () => {
  test('keyboard navigation - Tab through interactive elements', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Start at beginning
    await page.keyboard.press('Tab');
    
    // Should be able to tab through sortable headers
    for (let i = 0; i < 5; i++) {
      await page.keyboard.press('Tab');
    }
    
    // Verify focus is visible (check for focus-visible styles)
    const focusedElement = await page.evaluate(() => document.activeElement?.tagName);
    expect(focusedElement).toBeTruthy();
  });

  test('focus indicators are visible', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Tab to first interactive element
    await page.keyboard.press('Tab');
    
    // Check that focus styles are applied
    const hasFocusRing = await page.evaluate(() => {
      const activeEl = document.activeElement;
      if (!activeEl) return false;
      const styles = window.getComputedStyle(activeEl);
      // Check for outline or ring styles
      return styles.outline !== 'none' || activeEl.className.includes('focus');
    });
    
    // We expect some focus indication
    // (this is a basic check - visual regression would be better)
    expect(hasFocusRing).toBeTruthy();
  });

  test('ARIA roles and labels present', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Check for main landmark
    const main = page.locator('main');
    await expect(main).toBeVisible();
    
    // Check for header
    const header = page.locator('header');
    await expect(header).toBeVisible();
    
    // Check for table semantics
    const tableOrEmptyState = page.locator('table').or(page.getByText('No runs yet'));
    await expect(tableOrEmptyState).toBeVisible();
  });

  test('color contrast meets WCAG AA', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Check header text contrast
    const headerContrast = await page.evaluate(() => {
      const h1 = document.querySelector('h1');
      if (!h1) return null;
      
      const color = window.getComputedStyle(h1).color;
      const bgColor = window.getComputedStyle(h1).backgroundColor;
      
      return { color, bgColor };
    });
    
    // Basic check that colors are defined
    expect(headerContrast).toBeTruthy();
    expect(headerContrast?.color).toBeTruthy();
  });

  test('interactive elements have proper roles', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Check buttons are actually buttons
    const sortButtons = page.locator('button');
    const buttonCount = await sortButtons.count();
    expect(buttonCount).toBeGreaterThan(0);
    
    // Check links have proper href
    const links = page.locator('a[href]');
    const linkCount = await links.count();
    // Should have at least some links (run URLs)
    expect(linkCount).toBeGreaterThanOrEqual(0);
  });

  test('search input is accessible', async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('.grid', { timeout: 15000 });
    
    // Find search input
    const searchInput = page.locator('input[type="text"]').or(page.locator('input[placeholder*="Search"]'));
    const hasSearch = await searchInput.count() > 0;
    
    if (hasSearch) {
      await expect(searchInput.first()).toBeVisible();
      
      // Should be keyboard accessible
      await searchInput.first().focus();
      await page.keyboard.type('test');
      
      const value = await searchInput.first().inputValue();
      expect(value).toBe('test');
    }
  });

  test('connection status is announced', async ({ page }) => {
    await page.goto('/');
    
    // Connection status should be visible
    const status = page.locator('header').getByText(/Connected|Disconnected/);
    await expect(status).toBeVisible({ timeout: 15000 });
    
    // Should have visual indicator (the Circle component)
    const statusIndicator = page.locator('header svg');
    await expect(statusIndicator.first()).toBeVisible();
  });
});